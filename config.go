package main

import (
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	AurHelper string   `toml:"aurHelper"`
	ExtraArgs []string `toml:"extraArgs"`
}

var (
	PACMAN       = "pacman"
	DEFAULT_ARGS = []string{"-Syu"}
)

var defaultConfig = Config{}

func LoadConfig() *Config {
	path, err := xdg.ConfigFile("ditto/config.toml")
	if err != nil {
		panic(fmt.Errorf("could not find config file: %w", err))
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		saveConfig(path, &defaultConfig)
		return &defaultConfig
	} else if err != nil {
		panic(fmt.Errorf("could not read config: %w", err))
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		saveConfig(path, &defaultConfig)
		return &defaultConfig
	}

	return &cfg
}

func saveConfig(path string, cfg *Config) {
	data, err := toml.Marshal(cfg)
	if err != nil {
		panic(fmt.Errorf("could not marshal config: %w", err))
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		panic(fmt.Errorf("could not write config: %w", err))
	}
}

func (c *Config) PackageManager() string {
	if c.AurHelper != "" {
		return c.AurHelper
	}
	return PACMAN
}

func (c *Config) Args() []string {
	merged := make([]string, 0, len(DEFAULT_ARGS)+len(c.ExtraArgs))
	merged = append(merged, DEFAULT_ARGS...)
	merged = append(merged, c.ExtraArgs...)

	return merged
}
