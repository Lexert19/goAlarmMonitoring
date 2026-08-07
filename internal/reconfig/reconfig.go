package reconfig

import (
	"fmt"
	"strconv"
	"strings"

	"goAlarmMonitoring/internal/config"
	"goAlarmMonitoring/internal/sensor"
)

type Reconfigurator struct {
	sensors []sensor.Sensor
	cfg     *config.Config
}

func NewReconfigurator(sensors []sensor.Sensor, cfg *config.Config) *Reconfigurator {
	return &Reconfigurator{
		sensors: sensors,
		cfg:     cfg,
	}
}

func (r *Reconfigurator) ProcessCommand(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	if line == "list" || line == "ls" {
		fmt.Println("\n--- Active Sensors ---")
		for _, s := range r.sensors {
			fmt.Printf("ID: %s\n", s.ID())
		}
		fmt.Println("----------------------")
		return
	}

	parts := strings.Fields(line)
	if len(parts) != 4 || parts[0] != "reconfig" {
		fmt.Println("Invalid command. Usage: reconfig <sensor_id_prefix> <key> <value> (or 'list')")
		return
	}

	sensorIDInput := parts[1]
	key := parts[2]
	val := parts[3]

	var target sensor.Sensor
	foundCount := 0

	for _, s := range r.sensors {
		idStr := s.ID().String()
		if idStr == sensorIDInput || strings.HasPrefix(idStr, sensorIDInput) {
			target = s
			foundCount++
		}
	}

	if foundCount == 0 {
		fmt.Printf("Sensor matching prefix '%s' not found\n", sensorIDInput)
		return
	}
	if foundCount > 1 {
		fmt.Printf("Multiple sensors match prefix '%s'. Please provide more characters.\n", sensorIDInput)
		return
	}

	var value interface{}
	switch key {
	case "motion", "door", "smoke":
		f, err := strconv.ParseFloat(val, 64)
		if err != nil || f < 0 || f > 1 {
			fmt.Println("Value must be float 0-1")
			return
		}
		value = f
	case "ticker":
		i, err := strconv.Atoi(val)
		if err != nil || i <= 0 {
			fmt.Println("Value must be positive integer")
			return
		}
		value = i
	default:
		fmt.Println("Unknown key. Available: motion, door, smoke, ticker")
		return
	}

	if err := target.ReconfigureOne(key, value); err != nil {
		fmt.Printf("Reconfigure error: %v\n", err)
	} else {
		fmt.Printf("Sensor %s reconfigured successfully\n", target.ID())
	}
}
