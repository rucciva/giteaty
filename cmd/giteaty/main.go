package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	err := command.Run(ctx, os.Args)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
