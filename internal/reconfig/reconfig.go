package reconfig

import (
	"fmt"
	"strconv"
	"strings"

	"goAlarmMonitoring/internal/config"
	"goAlarmMonitoring/internal/sensor"

	"github.com/google/uuid"
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
	parts := strings.Fields(line)
	if len(parts) != 4 || parts[0] != "reconfig" {
		fmt.Println("Invalid command. Usage: reconfig <sensor_id> <key> <value>")
		return
	}
	sensorID, err := uuid.Parse(parts[1])
	if err != nil {
		fmt.Printf("Invalid sensor ID: %v\n", err)
		return
	}
	key := parts[2]
	val := parts[3]

	var target sensor.Sensor
	found := false
	for _, s := range r.sensors {
		if s.ID() == sensorID {
			target = s
			found = true
			break
		}
	}
	if !found {
		fmt.Printf("Sensor %s not found\n", sensorID)
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
		fmt.Printf("Sensor %s reconfigured\n", sensorID)
	}
}
