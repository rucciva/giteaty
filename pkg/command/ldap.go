package command

import (
	ldaps "github.com/nmcclain/ldap"
	"github.com/rucciva/giteaty/pkg/gitea"
	"github.com/rucciva/giteaty/pkg/gitea/globals"
	"github.com/rucciva/giteaty/pkg/ldaphandler"
	"github.com/urfave/cli/v2"
)

const (
	flagLDAPBaseDn            = "ldap-base-dn"
	flagLDAPSearchers         = "ldap-searchers"
	flagLDAPCacheSize         = "ldap-cache-size"
	flagLDAPCacheExpireSecond = "ldap-cache-expire-second"
	flagLDAPListenAddr        = "ldap-listen-addr"
)

func ldapFlag() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    flagLDAPBaseDn,
			EnvVars: []string{"LDAP_BASE_DN"},
			Value:   "dc=domain,dc=com",
		},
		&cli.StringSliceFlag{
			Name:    flagLDAPSearchers,
			EnvVars: []string{"LDAP_SEARCHERS"},
			Usage:   "gitea usernames allowed for ldap searching",
			Value:   cli.NewStringSlice("admin"),
		},
		&cli.IntFlag{
			Name:    flagLDAPCacheSize,
			EnvVars: []string{"LDAP_CACHE_SIZE"},
			Value:   1024 * 1024 * 1024,
		},
		&cli.IntFlag{
			Name:    flagLDAPCacheExpireSecond,
			EnvVars: []string{"LDAP_CACHE_EXPIRE_SECOND"},
			Value:   60,
		},
		&cli.StringFlag{
			Name:    flagLDAPListenAddr,
			EnvVars: []string{"LDAP_LISTEN_ADDR"},
			Value:   ":389",
		},
	}
}

func (cmd command) NewLDAPHandler(c *cli.Context, m gitea.Models) (ldaphandler.Interface, error) {
	return ldaphandler.New(
		ldaphandler.WithBaseDN(c.String(flagLDAPBaseDn)),
		ldaphandler.WithSearchers(c.StringSlice(flagLDAPSearchers)),
		ldaphandler.WithCache(c.Int(flagLDAPCacheSize), c.Int(flagLDAPCacheExpireSecond)),
		ldaphandler.WithModels(m),
	)
}

func (cmd command) StartLDAP(c *cli.Context) (err error) {
	h, err := cmd.NewLDAPHandler(c, globals.Models())
	if err != nil {
		return
	}

	quit := make(chan bool)
	go func() { <-c.Context.Done(); close(quit) }()

	s := ldaps.NewServer()
	s.EnforceLDAP = true
	s.BindFunc("", h)
	s.SearchFunc("", h)
	s.QuitChannel(quit)

	return s.ListenAndServe(c.String(flagLDAPListenAddr))
}
