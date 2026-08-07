package main

import (
	"context"
	"fmt"
	"goAlarmMonitoring/internal/alarm"
	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/internal/config"
	"goAlarmMonitoring/internal/logger"
	"goAlarmMonitoring/internal/reconfig"
	"goAlarmMonitoring/internal/sensor"
	"os"
	"os/signal"
	"syscall"
)

// reconfig 019fdb37-1b48-73fb-85f5-ebe7e3f3d9a4 ticker 1
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	eventBus := bus.NewEventBus()
	log := logger.NewLogger(eventBus)
	alarmSvc := alarm.NewAlarmService(eventBus)

	var sensors []sensor.Sensor

	if cfg.AutoStart {
		for i := 0; i < 1; i++ {
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

	reconfigurator := reconfig.NewReconfigurator(sensors, cfg)
	manual.SetReconfigurator(reconfigurator)

	log.Start(ctx)
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
	log.Stop()
	eventBus.Close()
}
