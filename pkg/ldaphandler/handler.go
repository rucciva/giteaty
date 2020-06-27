package ldaphandler

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"strings"

	"code.gitea.io/gitea/models"
	"github.com/coocood/freecache"
	"github.com/rucciva/giteaty/pkg/gitea"

	"github.com/nmcclain/ldap"
)

type option = func(h *handler) error

func Options() []option {
	return make([]func(h *handler) error, 0, 4)
}

func WithBaseDN(baseDN string) option {
	return func(h *handler) (err error) {
		h.baseDN = newNames(baseDN)
		return
	}
}

func WithSearchers(usernames []string) option {
	return func(h *handler) (err error) {
		for _, u := range usernames {
			h.searchers[u] = true
		}
		return
	}
}

func WithCache(size, expireSecond int) option {
	return func(h *handler) (err error) {
		h.cache = freecache.NewCache(size)
		h.cacheExpire = expireSecond
		return
	}
}

func WithModels(m gitea.Models) option {
	return func(h *handler) (err error) {
		h.models = m
		return
	}
}

type handler struct {
	baseDN names

	userParentRDN names
	userUAttr     string

	groupParentRDN names
	groupUAttr     string

	searchers map[string]bool

	cache       *freecache.Cache
	cacheExpire int

	models gitea.Models
}

var keyUsers = []byte("users")

// New return ldap's Binder, Searcher, & Closer
func New(opts ...option) (h *handler, err error) {
	h = &handler{
		baseDN:        newNames("dc=domain,dc=com"),
		userParentRDN: newNames("ou=users"),
		userUAttr:     "uid",

		groupParentRDN: newNames("ou=groups"),
		groupUAttr:     "cn",

		searchers: map[string]bool{"admin": true},
	}
	for _, opt := range opts {
		if err = opt(h); err != nil {
			return
		}
	}
	for u := range h.searchers {
		delete(h.searchers, u)
		h.searchers[strings.ToLower(h.getUserDN(u))] = true
	}
	return
}

func getRDN(DN string, parentsRDN ...string) (rDN string, err error) {
	baseDN := strings.ToLower("," + strings.Join(parentsRDN, ","))
	DN = strings.ToLower(DN)
	if !strings.HasSuffix(DN, baseDN) {
		return "", fmt.Errorf("Invalid DN: '%s' is not under '%s'", DN, baseDN)
	}
	return strings.TrimSuffix(DN, baseDN), nil
}

func (h *handler) Bind(bindDN, pw string, conn net.Conn) (res ldap.LDAPResultCode, err error) {
	rdn, err := getRDN(bindDN, h.userParentRDN.String(), h.baseDN.String())
	if err != nil {
		return ldap.LDAPResultInvalidDNSyntax, nil
	}
	parts := strings.Split(rdn, ",")
	if len(parts) != 1 {
		return ldap.LDAPResultInvalidDNSyntax, nil
	}
	uname := strings.TrimPrefix(parts[0], h.userUAttr+"=")
	if uname == parts[0] {
		return ldap.LDAPResultInvalidDNSyntax, nil
	}

	if _, err = h.models.UserSignIn(uname, pw); err != nil {
		return ldap.LDAPResultInvalidCredentials, nil
	}
	return ldap.LDAPResultSuccess, nil
}

func (h *handler) getUserDN(username string) string {
	return fmt.Sprintf("%s=%s,%s,%s", h.userUAttr, username, h.userParentRDN, h.baseDN)
}

func (h *handler) getTeamDN(org, team string) string {
	return fmt.Sprintf("%s=%s[%s],%s,%s", h.groupUAttr, org, team, h.groupParentRDN, h.baseDN)
}

func (h *handler) getOrgDN(org string) string {
	return fmt.Sprintf("%s=%s,%s,%s", h.groupUAttr, org, h.groupParentRDN, h.baseDN)
}

func (h *handler) checkSearchPermission(boundDN string, searchReq ldap.SearchRequest) error {
	if len(boundDN) < 1 {
		return fmt.Errorf("Search Error: Anonymous BindDN not allowed")
	}
	if !strings.HasSuffix(strings.ToLower(searchReq.BaseDN), h.baseDN.String()) {
		return fmt.Errorf("Search Error: search BaseDN %s is not in our BaseDN %s", searchReq.BaseDN, h.baseDN)
	}
	if !h.searchers[strings.ToLower(boundDN)] {
		return fmt.Errorf("Search Error: BindDN '%s' is not permitted to search", boundDN)
	}
	return nil
}

func (h *handler) listUsers() (entries []*ldap.Entry, err error) {
	users, _, err := h.models.SearchUsers(&models.SearchUserOptions{})
	if err != nil {
		return
	}
	orgs, _, err := h.models.SearchUsers(&models.SearchUserOptions{Type: models.UserTypeOrganization})
	if err != nil {
		return
	}
	orgByID := map[int64]*models.User{}
	for _, org := range orgs {
		orgByID[org.ID] = org
	}

	for _, user := range users {
		teams, err := h.models.GetUserTeams(user.ID, models.ListOptions{})
		if err != nil {
			return nil, err
		}
		dn := h.getUserDN(user.Name)
		attrs := []*ldap.EntryAttribute{}
		attrs = append(attrs, &ldap.EntryAttribute{Name: h.userUAttr, Values: []string{user.Name}})
		attrs = append(attrs, &ldap.EntryAttribute{Name: "displayName", Values: []string{user.FullName}})
		if !user.KeepEmailPrivate {
			attrs = append(attrs, &ldap.EntryAttribute{Name: "mail", Values: []string{user.Email}})
		}
		var memberOf []string
		memberOfOrg := map[int64]bool{}
		for _, team := range teams {
			org, ok := orgByID[team.OrgID]
			if !ok {
				continue
			}
			if !memberOfOrg[team.OrgID] {
				memberOf = append(memberOf, h.getOrgDN(org.Name))
				memberOfOrg[team.OrgID] = true
			}
			memberOf = append(memberOf, h.getTeamDN(org.Name, team.Name))
		}
		if len(memberOf) > 0 {
			attrs = append(attrs, &ldap.EntryAttribute{Name: "memberOf", Values: memberOf})
		}
		attrs = append(attrs, &ldap.EntryAttribute{Name: "objectClass", Values: []string{"inetorgperson"}})
		attrs = append(attrs, h.userParentRDN.Attributes()...)
		attrs = append(attrs, h.baseDN.Attributes()...)
		entries = append(entries, &ldap.Entry{DN: dn, Attributes: attrs})
	}
	return
}

func (h *handler) listUsersCached() (entries []*ldap.Entry, err error) {
	if h.cache == nil {
		return h.listUsers()
	}

	v, err := h.cache.Get(keyUsers)
	if err == nil {
		err = gob.NewDecoder(bytes.NewReader(v)).Decode(&entries)
	}
	if err != nil {
		if entries, err = h.listUsers(); err != nil {
			return
		}
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(entries); err != nil {
			return entries, nil
		}
		_ = h.cache.Set(keyUsers, buf.Bytes(), h.cacheExpire) //TODO: log if error
		return
	}
	return
}

// Search return all gitea users and depends on server to filter it
// only handle 'inetorgperson'
func (h *handler) Search(boundDN string, searchReq ldap.SearchRequest, conn net.Conn) (res ldap.ServerSearchResult, err error) {

	if err := h.checkSearchPermission(boundDN, searchReq); err != nil {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultInsufficientAccessRights}, err
	}

	class, err := ldap.GetFilterObjectClass(searchReq.Filter)
	if err != nil {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError},
			fmt.Errorf("Search Error: error parsing filter: %s", searchReq.Filter)
	}
	switch strings.ToLower(class) {
	case "":
	case "inetorgperson":
	default:
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError},
			fmt.Errorf("Search Error: unhandled filter type: %s [%s]", class, searchReq.Filter)
	}

	entries, err := h.listUsersCached()
	if err != nil {
		return ldap.ServerSearchResult{ResultCode: ldap.LDAPResultOperationsError}, err
	}
	res = ldap.ServerSearchResult{
		Entries:    entries,
		Referrals:  []string{},
		Controls:   []ldap.Control{},
		ResultCode: ldap.LDAPResultSuccess,
	}

	return
}

func (h *handler) Close(boundDN string, conn net.Conn) (err error) {
	return nil
}
