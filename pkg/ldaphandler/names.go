package ldaphandler

import (
	"strings"

	"github.com/nmcclain/ldap"
)

type names struct {
	str   string
	attrs map[string][]string
}

func newNames(s string) (n names) {
	var sb strings.Builder
	n.attrs = make(map[string][]string)

	needComma := false
	for _, entry := range strings.Split(s, ",") {

		needPlus := false
		for _, attr := range strings.Split(entry, "+") {
			if nv := strings.Split(attr, "="); len(nv) == 2 {
				if needComma {
					sb.WriteString(",")
					needComma = false
				}
				if needPlus {
					sb.WriteString("+")
				}
				sb.WriteString(attr)
				needPlus = true

				n.attrs[nv[0]] = append(n.attrs[nv[0]], nv[1])
			}
		}
		if needPlus {
			needComma = true
		}
	}
	n.str = sb.String()
	return
}

func (a names) String() string {
	return a.str
}

func (a names) Attributes() (attrs []*ldap.EntryAttribute) {
	for k, v := range a.attrs {
		attrs = append(attrs, &ldap.EntryAttribute{Name: k, Values: v})
	}
	return
}
