package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"git.rucciva.one/rucciva/log"
	"git.rucciva.one/rucciva/log/impl/rzap"

	"github.com/rucciva/giteaty/pkg/command"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-sigCh
		cancel()
	}()

	log.RegisterP(rzap.NewPLogger())

	err := command.Run(ctx, os.Args)
	if err != nil {
		log.GetPGlobal().Error("run_failed").WithFields("error", err)
		os.Exit(1)
	}
}
