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
			scenario: "InvalidBaseDN",
			dn:       fmt.Sprintf("%s=%s,%s", h.userUAttr, "some-username", "dc=rucciva,dc=com"),
		},
		{
			scenario: "InvalidSubBaseDN",
			dn:       fmt.Sprintf("%s=%s,%s", h.userUAttr, "some-username", "dc=someone,dc=one"),
		},
		{
			scenario: "InvalidUserParentRDN",
			dn:       fmt.Sprintf("%s=%s,%s", h.userUAttr, "some-username", h.baseDN),
		},
		{
			scenario: "InvalidUserUniqueAttribute",
			dn:       fmt.Sprintf("id=%s,%s,%s", "some-username", h.userParentRDN, h.baseDN),
		},
		{
			scenario: "RootDNOnly",
			dn:       h.baseDN.String(),
		},
		{
			scenario: "RootDNOnlyButInvalid",
			dn:       "dc=rucciva,dc=com",
		},
		{
			scenario: "RootDNOnlyButSubInvalid",
			dn:       "dc=someone,dc=one",
		},
		{
			scenario: "NoUserID",
			dn:       fmt.Sprintf("%s,%s", h.userParentRDN, h.baseDN),
		},
		{
			scenario: "NotDirectChildOfUserParentRDN",
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
			scenario: "SuccessfulLogin",
			username: "rucciva",
			password: "password",
			user:     &models.User{Name: "placeholder"},
			err:      nil,
		},
		{
			scenario: "InvalidCredenntial",
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
