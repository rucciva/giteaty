package ldaphandler

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"code.gitea.io/gitea/models"
	"github.com/golang/mock/gomock"
	"github.com/nmcclain/ldap"
	"github.com/rucciva/giteaty/internal/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reflectEq struct {
	m interface{}
}

func (m reflectEq) Matches(x interface{}) bool {
	return reflect.DeepEqual(m.m, x)
}

func (reflectEq) String() string {
	return "reflect"
}

func toJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func TestSearch(t *testing.T) {
	h, err := New(WithCache(1024*1024*1024, 5))
	require.NoError(t, err)

	data := []struct {
		user  *models.User
		orgs  []*models.User
		teams []*models.Team

		entry *ldap.Entry
	}{
		{
			user: &models.User{ID: 0, Name: "user", Email: "user@domain.com", FullName: "user me", IsActive: true},
			orgs: []*models.User{
				{ID: 2, Name: "org"},
				{ID: 3, Name: "org1"},
			},
			teams: []*models.Team{
				{ID: 0, OrgID: 2, Name: "team"},
				{ID: 1, OrgID: 2, Name: "team1"},
				{ID: 2, OrgID: 3, Name: "team2"},
				{ID: 3, OrgID: 3, Name: "team3"},
			},

			entry: &ldap.Entry{
				DN: h.getUserDN("user"),
				Attributes: []*ldap.EntryAttribute{
					{Name: "uid", Values: []string{"user"}},
					{Name: "displayName", Values: []string{"user me"}},
					{Name: "mail", Values: []string{"user@domain.com"}},
					{Name: "loginDisabled", Values: []string{"false"}},
					{Name: "memberOf", Values: []string{
						h.getOrgDN("org"),
						h.getTeamDN("org", "team"),
						h.getTeamDN("org", "team1"),
						h.getOrgDN("org1"),
						h.getTeamDN("org1", "team2"),
						h.getTeamDN("org1", "team3"),
					}},
					{Name: "objectClass", Values: []string{"inetorgperson"}},
					{Name: "ou", Values: []string{"users"}},
					{Name: "dc", Values: []string{"domain", "com"}},
				},
			},
		},
		{
			user: &models.User{ID: 1, Name: "user1", Email: "user1@domain.com", FullName: "user1 me", IsActive: true},
			orgs: []*models.User{
				{ID: 4, Name: "org2"},
				{ID: 5, Name: "org3"},
			},
			teams: []*models.Team{
				{ID: 0, OrgID: 2, Name: "team"},
				{ID: 1, OrgID: 2, Name: "team1"},
				{ID: 2, OrgID: 3, Name: "team2"},
				{ID: 3, OrgID: 3, Name: "team3"},
				{ID: 4, OrgID: 4, Name: "team4"},
				{ID: 5, OrgID: 4, Name: "team5"},
				{ID: 6, OrgID: 5, Name: "team6"},
				{ID: 7, OrgID: 5, Name: "team7"},
			},

			entry: &ldap.Entry{
				DN: h.getUserDN("user1"),
				Attributes: []*ldap.EntryAttribute{
					{Name: "uid", Values: []string{"user1"}},
					{Name: "displayName", Values: []string{"user1 me"}},
					{Name: "mail", Values: []string{"user1@domain.com"}},
					{Name: "loginDisabled", Values: []string{"false"}},
					{Name: "memberOf", Values: []string{
						h.getOrgDN("org"),
						h.getTeamDN("org", "team"),
						h.getTeamDN("org", "team1"),
						h.getOrgDN("org1"),
						h.getTeamDN("org1", "team2"),
						h.getTeamDN("org1", "team3"),
						h.getOrgDN("org2"),
						h.getTeamDN("org2", "team4"),
						h.getTeamDN("org2", "team5"),
						h.getOrgDN("org3"),
						h.getTeamDN("org3", "team6"),
						h.getTeamDN("org3", "team7"),
					}},
					{Name: "objectClass", Values: []string{"inetorgperson"}},
					{Name: "ou", Values: []string{"users"}},
					{Name: "dc", Values: []string{"domain", "com"}},
				},
			},
		},
		{
			user: &models.User{ID: 6, Name: "user2", Email: "user2@domain.com", FullName: "user2 me", IsActive: true},

			entry: &ldap.Entry{
				DN: h.getUserDN("user2"),
				Attributes: []*ldap.EntryAttribute{
					{Name: "uid", Values: []string{"user2"}},
					{Name: "displayName", Values: []string{"user2 me"}},
					{Name: "mail", Values: []string{"user2@domain.com"}},
					{Name: "loginDisabled", Values: []string{"false"}},
					{Name: "objectClass", Values: []string{"inetorgperson"}},
					{Name: "ou", Values: []string{"users"}},
					{Name: "dc", Values: []string{"domain", "com"}},
				},
			},
		},
		{
			user: &models.User{ID: 7, Name: "user3", Email: "user3@domain.com", FullName: "user3 me", KeepEmailPrivate: true, IsActive: true},

			entry: &ldap.Entry{
				DN: h.getUserDN("user3"),
				Attributes: []*ldap.EntryAttribute{
					{Name: "uid", Values: []string{"user3"}},
					{Name: "displayName", Values: []string{"user3 me"}},
					{Name: "loginDisabled", Values: []string{"false"}},
					{Name: "objectClass", Values: []string{"inetorgperson"}},
					{Name: "ou", Values: []string{"users"}},
					{Name: "dc", Values: []string{"domain", "com"}},
				},
			},
		},
		{
			user: &models.User{ID: 8, Name: "user4", Email: "user4@domain.com", FullName: "user4 me", KeepEmailPrivate: true},

			entry: &ldap.Entry{
				DN: h.getUserDN("user4"),
				Attributes: []*ldap.EntryAttribute{
					{Name: "uid", Values: []string{"user4"}},
					{Name: "displayName", Values: []string{"user4 me"}},
					{Name: "loginDisabled", Values: []string{"true"}},
					{Name: "objectClass", Values: []string{"inetorgperson"}},
					{Name: "ou", Values: []string{"users"}},
					{Name: "dc", Values: []string{"domain", "com"}},
				},
			},
		},
	}
	users, groups, entries := []*models.User{}, []*models.User{}, []*ldap.Entry{}
	for _, dat := range data {
		users, groups, entries = append(users, dat.user), append(groups, dat.orgs...), append(entries, dat.entry)
	}
	cacheMissCount := 2

	// mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mdl := mock.NewMockModels(ctrl)
	mdl.EXPECT().
		SearchUsers(reflectEq{&models.SearchUserOptions{}}).
		Return(users, int64(len(users)), nil).Times(cacheMissCount)
	mdl.EXPECT().
		SearchUsers(reflectEq{&models.SearchUserOptions{Type: models.UserTypeOrganization}}).
		Return(groups, int64(len(groups)), nil).Times(cacheMissCount)
	for _, dat := range data {
		mdl.EXPECT().
			GetUserTeams(gomock.Eq(dat.user.ID), reflectEq{models.ListOptions{}}).
			Return(dat.teams, nil).Times(cacheMissCount)
	}
	h.models = mdl

	filter := "(&(objectClass=InetOrgPerson)(uid=*))"
	req := ldap.SearchRequest{BaseDN: h.baseDN.String(), Filter: filter}
	for i := 0; i < cacheMissCount; i++ {
		for j := 0; j < 3; j++ {
			res, err := h.Search(h.getUserDN("admin"), req, nil)
			require.NoError(t, err, "should not return error")
			assert.Equal(t, ldap.LDAPResultCode(ldap.LDAPResultSuccess), res.ResultCode)
			assert.ElementsMatchf(t, entries, res.Entries, "\n%s\nshould match\n%s", toJSONString(res.Entries), toJSONString(entries))
		}
		if i < cacheMissCount-1 {
			<-time.After(time.Duration(h.cacheExpire) * time.Second) // wait cache expire
		}
	}
}

func TestSearchInvalid(t *testing.T) {
	h, err := New()
	require.NoError(t, err)

	data := []struct {
		scenario string

		bindDN string
		baseDN string
		filter string

		result ldap.LDAPResultCode
	}{
		{
			scenario: "AnonymousBindDN",
			bindDN:   "",
			baseDN:   h.baseDN.String(),
			filter:   "(&(objectClass=InetOrgPerson)(uid=*))",
			result:   ldap.LDAPResultInsufficientAccessRights,
		},
		{
			scenario: "UnprivilegedBindDN",
			bindDN:   "uid=someuser,ou=users,dc=domain,dc=com",
			baseDN:   h.baseDN.String(),
			filter:   "(&(objectClass=InetOrgPerson)(uid=*))",
			result:   ldap.LDAPResultInsufficientAccessRights,
		},
		{
			scenario: "UnknownBaseDN",
			bindDN:   "uid=admin,ou=users,dc=domain,dc=com",
			baseDN:   "dc=domain,dc=net",
			filter:   "(&(objectClass=InetOrgPerson)(uid=*))",
			result:   ldap.LDAPResultInsufficientAccessRights,
		},
		{
			scenario: "UnknownSubBaseDN",
			bindDN:   "uid=admin,ou=users,dc=domain,dc=com",
			baseDN:   "dc=domain1,dc=com",
			filter:   "(&(objectClass=InetOrgPerson)(uid=*))",
			result:   ldap.LDAPResultInsufficientAccessRights,
		},
		{
			scenario: "UnknownObjectClass",
			bindDN:   "uid=admin,ou=users,dc=domain,dc=com",
			baseDN:   "dc=domain,dc=com",
			filter:   "(&(objectClass=person)(uid=*))",
			result:   ldap.LDAPResultOperationsError,
		},
	}

	for _, dat := range data {
		t.Run(dat.scenario, func(t *testing.T) {
			req := ldap.SearchRequest{BaseDN: dat.baseDN, Filter: dat.filter}
			res, err := h.Search(dat.bindDN, req, nil)
			assert.Error(t, err)
			assert.Equal(t, dat.result, res.ResultCode)
			assert.Empty(t, res.Entries)
		})
	}
}
