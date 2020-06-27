package command

import (
	"context"
	"fmt"

	cli "github.com/urfave/cli/v2"
)

type command struct {
}

func (cmd command) Action(c *cli.Context) (err error) {
	if err = cmd.InitDB(c); err != nil {
		return fmt.Errorf("init database failed: %v", err)
	}

	return cmd.StartLDAP(c)
}

func (cmd command) Run(ctx context.Context, args []string) (err error) {
	flags := append(modelsFlag(), ldapFlag()...)

	app := &cli.App{
		Name:   "giteaty",
		Flags:  flags,
		Action: cmd.Action,
	}

	return app.RunContext(ctx, args)
}

func Run(ctx context.Context, args []string) (err error) {
	return command{}.Run(ctx, args)
}
