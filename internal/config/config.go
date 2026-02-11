package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"geoswitch/internal/types"

	"go.yaml.in/yaml/v4"
)

// ExitConfig defines the configuration for a network exit point.
type ExitConfig struct {
	Provider string `yaml:"provider"`
	Country  string `yaml:"country"`
}

type Config struct {
	DefaultExit string                `yaml:"default_exit"`
	Exits       map[string]ExitConfig `yaml:"exits"`
}

// LoadConfig reads and parses a YAML configuration file.
func LoadConfig(path string) (*Config, error) {
	log.Printf("[config] loading configuration from %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[config] failed to read config file: %v", err)
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Printf("[config] failed to parse YAML: %v", err)
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Populate the Name field from map keys
	for name, exit := range config.Exits {
		name = strings.ToLower(name)
		config.Exits[name] = exit
	}

	if err := config.Validate(); err != nil {
		log.Printf("[config] validation failed: %v", err)
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	log.Printf("[config] successfully loaded config with default exit '%s' and %d exits", config.DefaultExit, len(config.Exits))
	return &config, nil
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.DefaultExit == "" {
		return fmt.Errorf("default_exit is required")
	}

	if len(c.Exits) == 0 {
		return fmt.Errorf("at least one exit must be defined")
	}

	if _, ok := c.Exits[c.DefaultExit]; !ok {
		return fmt.Errorf("default_exit '%s' is not defined in exits", c.DefaultExit)
	}

	for name, exit := range c.Exits {
		if exit.Provider == "" {
			return fmt.Errorf("exit '%s': provider is required", name)
		}
		if exit.Country == "" {
			return fmt.Errorf("exit '%s': country is required", name)
		}
	}

	return nil
}

type ConfigExitResolver struct {
	Config *Config
}

func (r *ConfigExitResolver) Resolve(exit *types.Exit) (string, ExitConfig, error) {
	// No exit specified → default
	if exit == nil || exit.Name == "" {
		name := r.Config.DefaultExit
		cfg, _ := r.Config.GetExit(name)
		log.Printf("[resolver] using default exit: %s", name)
		return name, cfg, nil
	}

	cfg, ok := r.Config.GetExit(exit.Name)
	if !ok {
		return "", ExitConfig{}, fmt.Errorf("unknown exit '%s'", exit.Name)
	}

	return exit.Name, cfg, nil
}

func (c *Config) GetExit(name string) (ExitConfig, bool) {
	exit, ok := c.Exits[name]
	return exit, ok
}

func (r *ConfigExitResolver) defaultExit() (ExitConfig, error) {
	cfg, ok := r.Config.GetExit(r.Config.DefaultExit)
	if !ok {
		// This should never happen if config was validated
		return ExitConfig{}, fmt.Errorf("default exit '%s' not defined", r.Config.DefaultExit)
	}

	return cfg, nil
}
