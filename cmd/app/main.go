package main

import (
	"context"
	"fmt"
	"goAlarmMonitoring/internal/alarm"
	"goAlarmMonitoring/internal/analysis"
	"goAlarmMonitoring/internal/bus"
	"goAlarmMonitoring/internal/config"
	"goAlarmMonitoring/internal/logger"
	"goAlarmMonitoring/internal/registry"
	"goAlarmMonitoring/internal/sensor"
	"os"
	"os/signal"
	"syscall"
)

// reconfig 019fdc08-1304-735a-a245-93cdf47ee709 ticker 1
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	eventBus := bus.NewEventBus()
	log := logger.NewLogger(eventBus)
	alarmRepo := alarm.NewMemoryAlarmRepository()
	alarmSvc := alarm.NewAlarmService(eventBus, alarmRepo)
	analysisSvc := analysis.NewAnalysisService(eventBus, alarmSvc)
	alarmLog := alarm.NewAlarmLogger(eventBus)

	devRegistry := registry.NewMemoryDeviceRegistry(eventBus, cfg)

	manual := sensor.NewManualSensor(eventBus, devRegistry)
	if err := devRegistry.Register(manual); err != nil {
		fmt.Println("Error registering manual sensor:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Start(ctx)
	analysisSvc.Start(ctx)
	alarmLog.Start(ctx)

	if err := manual.Start(ctx); err != nil {
		fmt.Println("Error starting manual sensor:", err)
	}
	fmt.Printf("Manual sensor started with ID: %s\n", manual.ID())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nShutting down...")
	cancel()
	devRegistry.StopAll()
	analysisSvc.Stop()
	alarmLog.Stop()
	log.Stop()
	eventBus.Close()
}
