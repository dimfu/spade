package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dimfu/spade/config"
	"github.com/dimfu/spade/discord"
)

func main() {
	config.Init()

	ctx, cancel := context.WithCancel(context.Background())

	go discord.Init(ctx)

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	<-stopChan
	log.Println("Shutting down...")

	cancel()

	log.Println("Shutdown complete")
}
