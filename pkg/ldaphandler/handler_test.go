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
		groupUAttr     string
		group          string
		team           string

		orgDN  string
		teamDN string
	}{
		{

			groupUAttr:     "cn",
			group:          "group",
			team:           "team",
			groupParentRDN: "ou=groups",
			baseDN:         "dc=domain,dc=com",
			teamDN:         "cn=group[team],ou=groups,dc=domain,dc=com",
			orgDN:          "cn=group,ou=groups,dc=domain,dc=com",
		},
	}

	for _, d := range data {
		h := handler{
			baseDN:         newNames(d.baseDN),
			groupParentRDN: newNames(d.groupParentRDN),
			groupUAttr:     d.groupUAttr,
		}
		assert.Equal(t, d.teamDN, h.getTeamDN(d.group, d.team))
		assert.Equal(t, d.orgDN, h.getOrgDN(d.group))
	}
}
