package ldaphandler

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"code.gitea.io/gitea/models"
	ldapc "github.com/go-ldap/ldap/v3"
	"github.com/nmcclain/ldap"
	"github.com/rucciva/giteaty/pkg/gitea/globals"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tCreateUser(u models.User) (user *models.User, closer func(), err error) {
	user = &u
	if err = models.CreateUser(user); err != nil {
		return
	}
	closer = func() { _ = models.DeleteUser(user) }
	return
}

type tLDAPServer struct {
	handler Interface

	quit chan bool
	done bool
	port uint16
	err  error
}

func newLDAPTestServer(t *testing.T, handler Interface) (s *tLDAPServer, err error) {
	s = &tLDAPServer{
		quit:    make(chan bool),
		handler: handler,
		port:    uint16(rand.Intn(65535-1024) + 1024),
	}
	go s.start(t)
	err = s.waitConnection()
	return
}

func (s *tLDAPServer) start(t *testing.T) {
	sv := ldap.NewServer()
	sv.EnforceLDAP = true
	sv.BindFunc("", s.handler)
	sv.SearchFunc("", s.handler)
	sv.QuitChannel(s.quit)
	for {
		t.Logf("listenting on :%d", s.port)
		s.err = sv.ListenAndServe(fmt.Sprintf(":%d", s.port))
		if s.done {
			t.Logf("closed")
			return
		}
		t.Logf("listen error: %v", s.err)
		s.port += 1
	}
}

func (s *tLDAPServer) waitConnection() (err error) {
	maxDelay := time.Second
	delay := time.Millisecond
	var conn net.Conn
	for delay <= maxDelay {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf(":%d", s.port), time.Second)
		if conn != nil {
			conn.Close()
		}
		if err == nil {
			return
		}
		<-time.After(delay)
		delay = delay * 2
	}
	return
}

func (s *tLDAPServer) url() string {
	return fmt.Sprintf("ldap://0.0.0.0:%d", s.port)
}

func (s *tLDAPServer) close() error {
	s.done = true
	close(s.quit)
	return nil
}

func TestLDAP(t *testing.T) {
	if NoDB() {
		t.Skip()
	}
	users := []models.User{
		{
			Name:     "rucciva",
			Passwd:   "rucciva",
			Email:    "rucciva@gmail.com",
			FullName: "rucciva",
		},
		{
			Name:     "other",
			Passwd:   "other",
			Email:    "other@gmail.com",
			FullName: "other",
		},
		{
			Name:     "other1",
			Passwd:   "other1",
			Email:    "other1@gmail.com",
			FullName: "other1",
		},
	}
	for _, user := range users {
		_, closer, err := tCreateUser(user)
		require.NoError(t, err, "should not return error")
		defer closer()
	}
	admin, other := users[0], users[1]

	// instantiate server
	handler, err := New(
		WithBaseDN("dc=domain,dc=com"),
		WithSearchers([]string{admin.Name}),
		WithCache(2^10, 60),
		WithModels(globals.Models()),
	)
	require.NoError(t, err, "should not return error")
	require.NotNil(t, handler, "handler should not be nil")
	s, err := newLDAPTestServer(t, handler)
	require.NoError(t, err, "should be able to start ldap server")
	defer s.close()

	// Login Flow
	l, err := ldapc.DialURL(s.url())
	require.NoError(t, err, "should be contactable")
	defer l.Close()

	err = l.Bind(handler.getUserDN(admin.Name), admin.Passwd)
	require.NoError(t, err, "should not return login error")

	req := ldapc.NewSearchRequest(
		handler.baseDN.String(),
		ldapc.ScopeWholeSubtree, ldapc.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=inetorgperson)(uid=%s)(ou=users)(dc=domain)(dc=com))", other.Name),
		[]string{}, nil,
	)
	res, err := l.Search(req)
	require.NoError(t, err, "should not return search error")
	require.Len(t, res.Entries, 1, "should return only one user")
	assert.Equal(t, res.Entries[0].DN, handler.getUserDN(other.Name))

	err = l.Bind(res.Entries[0].DN, other.Passwd)
	require.NoError(t, err, "should be able to login")
}
