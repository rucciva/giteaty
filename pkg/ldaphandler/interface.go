package ldaphandler

import "github.com/nmcclain/ldap"

type Interface interface {
	ldap.Binder
	ldap.Searcher
	ldap.Closer
}

var (
	_ Interface = &handler{}
)
