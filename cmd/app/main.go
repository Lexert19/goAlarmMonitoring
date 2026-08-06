package main

import (
	"context"
	"fmt"
	"goAlarmMonitoring/internal/alarm"
	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/internal/config"
	"goAlarmMonitoring/internal/logger"
	"goAlarmMonitoring/internal/sensor"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	eventBus := bus.NewEventBus(cfg.BufferSize)
	logger := logger.NewLogger(eventBus)
	alarmSvc := alarm.NewAlarmService(eventBus)

	var sensors []sensor.Sensor

	if cfg.AutoStart {
		for i := 0; i < 2; i++ {
			s := sensor.NewSimulatedSensor(eventBus, cfg)
			sensors = append(sensors, s)
			fmt.Printf("Starting sensor %d with ID: %s\n", i+1, s.ID())
		}
	}

	manual := sensor.NewManualSensor(eventBus)
	sensors = append(sensors, manual)
	fmt.Printf("Manual sensor started with ID: %s\n", manual.ID())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Start(ctx)
	alarmSvc.Start(ctx)

	for _, s := range sensors {
		if err := s.Start(ctx); err != nil {
			fmt.Printf("Error starting sensor %s: %v\n", s.ID(), err)
		}
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	cancel()
	for _, s := range sensors {
		s.Stop()
	}
	alarmSvc.Stop()
	eventBus.Close()
}
