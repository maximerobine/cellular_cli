package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	app, err := NewApplication(log)
	if err != nil {
		panic(err)
	}
	errCh := app.Start()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	select {
	case signalErr := <-signalCh:
		log.Info(fmt.Sprintf("system interruption signal received: %s", signalErr))
		app.Stop()
	case err := <-errCh:
		log.Error(fmt.Sprintf("error while running tge application: %v", err))
	}
}
