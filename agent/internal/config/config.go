package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ServerID        string   `mapstructure:"server_id"`
	HQAddress       string   `mapstructure:"hq_address"`
	IntervalSeconds int      `mapstructure:"interval_seconds"`
	Services        []string `mapstructure:"services"`
}

func Load(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// Also allow environment variable overrides
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}