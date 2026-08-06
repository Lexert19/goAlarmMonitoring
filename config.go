package main

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AutoStart  bool    `yaml:"auto_start" env:"AUTO_START"`
	TickerSec  int     `yaml:"ticker_sec" env:"TICKER_SEC"`
	MotionProb float64 `yaml:"motion_prob" env:"MOTION_PROB"`
	DoorProb   float64 `yaml:"door_prob" env:"DOOR_PROB"`
	SmokeProb  float64 `yaml:"smoke_prob" env:"SMOKE_PROB"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		AutoStart:  true,
		TickerSec:  2,
		MotionProb: 0.3,
		DoorProb:   0.2,
		SmokeProb:  0.1,
	}

	if err := cleanenv.ReadConfig("config.yaml", cfg); err != nil {
	}

	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
