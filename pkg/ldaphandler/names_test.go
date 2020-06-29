package ldaphandler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNames(t *testing.T) {
	data := []struct {
		i, o string
	}{
		{i: "", o: ""},
		{i: "dc", o: ""},
		{i: "dc=", o: "dc="},
		{i: "dc=com", o: "dc=com"},
		{i: "dc=com,,,,", o: "dc=com"},
		{i: ",,,,,dc=com", o: "dc=com"},
		{i: "dc,dc=com", o: "dc=com"},
		{i: "dc=domain,dc,dc=com", o: "dc=domain,dc=com"},
		{i: "dc=domain,dc,dc,dc=com", o: "dc=domain,dc=com"},
		{i: "dc=domain,,,,,dc=com", o: "dc=domain,dc=com"},
		{i: "dc=domain,dc=com", o: "dc=domain,dc=com"},
		{i: "cn=rucciva+cn=noneedtoknow", o: "cn=rucciva+cn=noneedtoknow"},
		{i: "cn=rucciva+++++", o: "cn=rucciva"},
		{i: "++++cn=rucciva", o: "cn=rucciva"},
		{i: "cn=rucciva+cn=noneedtoknow,dc=domain,dc=com", o: "cn=rucciva+cn=noneedtoknow,dc=domain,dc=com"},
		{i: "cn=rucciva++++cn=noneedtoknow,,,,,dc=domain,dc=com", o: "cn=rucciva+cn=noneedtoknow,dc=domain,dc=com"},
		{i: "cn=rucciva++,,++cn=noneedtoknow,,,,,dc=domain,dc=com", o: "cn=rucciva,cn=noneedtoknow,dc=domain,dc=com"},
		{i: "cn=rucciva++,,++cn=noneedtoknow,,++,,,dc=domain,dc=com", o: "cn=rucciva,cn=noneedtoknow,dc=domain,dc=com"},
	}

	for _, d := range data {
		assert.Equalf(t, d.o, newNames(d.i).String(), "input: %s", d.i)
	}
}
