package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	bus := NewEventBus()
	logger := NewLogger()
	alarm := NewAlarmService(bus, logger)
	sensor := NewSensor(bus, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	alarm.Start(ctx)

	if cfg.AutoStart {
		if err := sensor.Start(ctx); err != nil {
			fmt.Println("Error starting sensor:", err)
		}
	}

	go ManualInputLoop(ctx, bus)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	cancel()
	sensor.Stop()
	alarm.Stop()
	bus.Close()
}
