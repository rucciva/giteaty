package command

import (
	"context"
	"fmt"

	"github.com/rucciva/giteaty/pkg/gitea/globals"
	cli "github.com/urfave/cli/v2"
)

func action(c *cli.Context) (err error) {
	if err = initDB(c); err != nil {
		return fmt.Errorf("init database failed: %v", err)
	}

	return startLDAP(c, globals.Models())
}

func Run(ctx context.Context, args []string) (err error) {
	flags := append(modelsFlag(), ldapFlag()...)
	app := &cli.App{
		Name:   "giteaty",
		Flags:  flags,
		Action: action,
	}
	return app.RunContext(ctx, args)
}
