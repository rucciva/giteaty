package ldaphandler

import (
	"fmt"
	"testing"

	"code.gitea.io/gitea/models"
	"github.com/golang/mock/gomock"
	"github.com/nmcclain/ldap"
	"github.com/rucciva/giteaty/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindInvalidDN(t *testing.T) {
	h, err := New()
	require.NoError(t, err)

	data := []struct {
		scenario string
		dn       string
	}{
		{
			scenario: "invalid BaseDN",
			dn:       fmt.Sprintf("%s=%s,%s", h.userUAttr, "some-username", "dc=rucciva,dc=com"),
		},
		{
			scenario: "invalid sub-BaseDN",
			dn:       fmt.Sprintf("%s=%s,%s", h.userUAttr, "some-username", "dc=someone,dc=one"),
		},
		{
			scenario: "invalid user parent RDN",
			dn:       fmt.Sprintf("%s=%s,%s", h.userUAttr, "some-username", h.baseDN),
		},
		{
			scenario: "invalid user unique attribute",
			dn:       fmt.Sprintf("id=%s,%s,%s", "some-username", h.userParentRDN, h.baseDN),
		},
		{
			scenario: "Root DN Only",
			dn:       h.baseDN.String(),
		},
		{
			scenario: "Root DN Only but Invalid",
			dn:       "dc=rucciva,dc=com",
		},
		{
			scenario: "Root DN Only but Sub Invalid",
			dn:       "dc=someone,dc=one",
		},
		{
			scenario: "No user ID",
			dn:       fmt.Sprintf("%s,%s", h.userParentRDN, h.baseDN),
		},
		{
			scenario: "Not direct child of userParentRDN",
			dn:       fmt.Sprintf("%s=%s,org=test,%s,%s", h.userUAttr, "some-username", h.userParentRDN, h.baseDN),
		},
	}

	for _, dat := range data {
		t.Run(dat.scenario, func(t *testing.T) {
			res, err := h.Bind(dat.dn, "some-pass", nil)
			assert.NoError(t, err, "should not return error")
			assert.Equal(t, ldap.LDAPResultCode(ldap.LDAPResultInvalidDNSyntax), res, "should return invalid dn syntax")
		})
	}
}

func TestBind(t *testing.T) {
	h, err := New()
	require.NoError(t, err)

	data := []struct {
		scenario string

		username string
		password string

		user *models.User
		err  error
	}{
		{
			scenario: "successful login",
			username: "rucciva",
			password: "password",
			user:     &models.User{Name: "placeholder"},
			err:      nil,
		},
		{
			scenario: "invalid credenntial",
			username: "rucciva",
			password: "invalid",
			user:     nil,
			err:      fmt.Errorf("user not exist"),
		},
	}

	for _, dat := range data {
		t.Run(dat.scenario, func(t *testing.T) {
			if dat.user != nil {
				dat.user.Name = dat.username
			}

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mdl := mock.NewMockModels(ctrl)
			mdl.EXPECT().
				UserSignIn(gomock.Eq(dat.username), gomock.Eq(dat.password)).
				Return(dat.user, dat.err).Times(1)
			h.models = mdl

			dn := h.getUserDN(dat.username)
			res, err := h.Bind(dn, dat.password, nil)
			assert.NoError(t, err)
			switch {
			case dat.user != nil:
				assert.Equal(t, ldap.LDAPResultCode(ldap.LDAPResultSuccess), res)
			default:
				assert.Equal(t, ldap.LDAPResultCode(ldap.LDAPResultInvalidCredentials), res)
			}
		})

	}
}
