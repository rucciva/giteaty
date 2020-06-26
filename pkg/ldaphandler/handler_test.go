package ldaphandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserDN(t *testing.T) {
	data := []struct {
		baseDN        string
		userParentRDN string
		userUAttr     string
		user          string

		dn string
	}{
		{
			userUAttr:     "uid",
			user:          "test",
			userParentRDN: "ou=users",
			baseDN:        "dc=domain,dc=com",
			dn:            "uid=test,ou=users,dc=domain,dc=com",
		},
		{
			userUAttr:     "uid",
			user:          "test",
			userParentRDN: "ou=users",
			baseDN:        "dc=domain,dc=com",
			dn:            "uid=test,ou=users,dc=domain,dc=com",
		},
	}

	for _, d := range data {
		h := handler{
			baseDN:        newNames(d.baseDN),
			userParentRDN: newNames(d.userParentRDN),
			userUAttr:     d.userUAttr,
		}
		assert.Equal(t, d.dn, h.getUserDN(d.user))
	}
}

func TestGetGroupDN(t *testing.T) {
	data := []struct {
		baseDN         string
		groupParentRDN string
		groupOrgUAttr  string
		groupTeamUAttr string
		group          string
		team           string

		dn string
	}{
		{

			groupTeamUAttr: "cn",
			team:           "team",
			groupOrgUAttr:  "cn",
			group:          "group",
			groupParentRDN: "ou=groups",
			baseDN:         "dc=domain,dc=com",
			dn:             "cn=team,cn=group,ou=groups,dc=domain,dc=com",
		},
	}

	for _, d := range data {
		h := handler{
			baseDN:         newNames(d.baseDN),
			groupParentRDN: newNames(d.groupParentRDN),
			groupOrgUAttr:  d.groupOrgUAttr,
			groupTeamUAttr: d.groupTeamUAttr,
		}
		assert.Equal(t, d.dn, h.getGroupDN(d.group, d.team))
	}
}
