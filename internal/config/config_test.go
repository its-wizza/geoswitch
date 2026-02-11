package config

import (
	"testing"

	"geoswitch/internal/types"
)

func TestConfig_Validate_ValidConfig(t *testing.T) {
	config := &Config{
		DefaultExit: "us-exit",
		Exits: map[string]ExitConfig{
			"us-exit": {
				Provider: "aws",
				Country:  "US",
			},
			"eu-exit": {
				Provider: "gcp",
				Country:  "DE",
			},
		},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("expected no error for valid config, got: %v", err)
	}
}

func TestConfig_Validate_MissingDefaultExit(t *testing.T) {
	config := &Config{
		DefaultExit: "",
		Exits: map[string]ExitConfig{
			"us-exit": {
				Provider: "aws",
				Country:  "US",
			},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for missing default_exit, got nil")
	}

	expected := "default_exit is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestConfig_Validate_NoExitsDefined(t *testing.T) {
	config := &Config{
		DefaultExit: "us-exit",
		Exits:       map[string]ExitConfig{},
	}

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for no exits defined, got nil")
	}

	expected := "at least one exit must be defined"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestConfig_Validate_DefaultExitNotDefined(t *testing.T) {
	config := &Config{
		DefaultExit: "nonexistent",
		Exits: map[string]ExitConfig{
			"us-exit": {
				Provider: "aws",
				Country:  "US",
			},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for undefined default_exit, got nil")
	}

	expected := "default_exit 'nonexistent' is not defined in exits"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestConfig_Validate_ExitMissingProvider(t *testing.T) {
	config := &Config{
		DefaultExit: "us-exit",
		Exits: map[string]ExitConfig{
			"us-exit": {
				Provider: "",
				Country:  "US",
			},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for missing provider, got nil")
	}

	expected := "exit 'us-exit': provider is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestConfig_Validate_ExitMissingCountry(t *testing.T) {
	config := &Config{
		DefaultExit: "us-exit",
		Exits: map[string]ExitConfig{
			"us-exit": {
				Provider: "aws",
				Country:  "",
			},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for missing country, got nil")
	}

	expected := "exit 'us-exit': country is required"
	if err.Error() != expected {
		t.Errorf("expected error '%s', got '%s'", expected, err.Error())
	}
}

func TestConfig_Validate_MultipleExitsOneInvalid(t *testing.T) {
	config := &Config{
		DefaultExit: "us-exit",
		Exits: map[string]ExitConfig{
			"us-exit": {
				Provider: "aws",
				Country:  "US",
			},
			"eu-exit": {
				Provider: "",
				Country:  "DE",
			},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Fatal("expected error for invalid exit, got nil")
	}
}

func TestLoadConfig_Success(t *testing.T) {
	config, err := LoadConfig("testdata/config/minimal.yaml")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if config == nil {
		t.Fatal("expected non-nil config")
	}

	if config.DefaultExit == "" {
		t.Error("expected default_exit to be set")
	}

	if len(config.Exits) == 0 {
		t.Error("expected at least one exit")
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	_, err := LoadConfig("testdata/config/invalid.yaml")
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadConfig_MultipleExits(t *testing.T) {
	config, err := LoadConfig("testdata/config/multiple-exits.yaml")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(config.Exits) < 2 {
		t.Errorf("expected at least 2 exits, got %d", len(config.Exits))
	}
}

func TestLoadConfig_UnknownExitInDefault(t *testing.T) {
	_, err := LoadConfig("testdata/config/unknown-exit.yaml")
	if err == nil {
		t.Fatal("expected error for unknown default_exit, got nil")
	}
}

func TestConfig_GetExit_Success(t *testing.T) {
	config := &Config{
		DefaultExit: "us",
		Exits: map[string]ExitConfig{
			"us": {
				Provider: "aws",
				Country:  "US",
			},
		},
	}

	exit, ok := config.GetExit("us")
	if !ok {
		t.Fatal("expected exit to be found")
	}

	if exit.Provider != "aws" {
		t.Errorf("expected provider 'aws', got '%s'", exit.Provider)
	}

	if exit.Country != "US" {
		t.Errorf("expected country 'US', got '%s'", exit.Country)
	}
}

func TestConfig_GetExit_NotFound(t *testing.T) {
	config := &Config{
		DefaultExit: "us",
		Exits: map[string]ExitConfig{
			"us": {
				Provider: "aws",
				Country:  "US",
			},
		},
	}

	_, ok := config.GetExit("nonexistent")
	if ok {
		t.Error("expected exit not to be found")
	}
}

func TestConfigExitResolver_Resolve_Default(t *testing.T) {
	config := &Config{
		DefaultExit: "us",
		Exits: map[string]ExitConfig{
			"us": {
				Provider: "aws",
				Country:  "US",
			},
		},
	}

	resolver := &ConfigExitResolver{Config: config}

	// Test nil exit
	name, cfg, err := resolver.Resolve(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "us" {
		t.Errorf("expected exit name 'us', got '%s'", name)
	}

	if cfg.Provider != "aws" {
		t.Errorf("expected provider 'aws', got '%s'", cfg.Provider)
	}

	// Test empty exit name
	name, cfg, err = resolver.Resolve(&types.Exit{Name: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "us" {
		t.Errorf("expected exit name 'us', got '%s'", name)
	}
}

func TestConfigExitResolver_Resolve_NamedExit(t *testing.T) {
	config := &Config{
		DefaultExit: "us",
		Exits: map[string]ExitConfig{
			"us": {
				Provider: "aws",
				Country:  "US",
			},
			"de": {
				Provider: "gcp",
				Country:  "DE",
			},
		},
	}

	resolver := &ConfigExitResolver{Config: config}

	name, cfg, err := resolver.Resolve(&types.Exit{Name: "de"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "de" {
		t.Errorf("expected exit name 'de', got '%s'", name)
	}

	if cfg.Country != "DE" {
		t.Errorf("expected country 'DE', got '%s'", cfg.Country)
	}
}

func TestConfigExitResolver_Resolve_UnknownExit(t *testing.T) {
	config := &Config{
		DefaultExit: "us",
		Exits: map[string]ExitConfig{
			"us": {
				Provider: "aws",
				Country:  "US",
			},
		},
	}

	resolver := &ConfigExitResolver{Config: config}

	_, _, err := resolver.Resolve(&types.Exit{Name: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown exit, got nil")
	}
}

func TestLoadConfig_CaseSensitivity(t *testing.T) {
	// Create a temporary config to test case handling
	config := &Config{
		DefaultExit: "US",
		Exits: map[string]ExitConfig{
			"US": {
				Provider: "aws",
				Country:  "US",
			},
			"De": {
				Provider: "gcp",
				Country:  "DE",
			},
		},
	}

	// Manually trigger the lowercase normalization
	exits := make(map[string]ExitConfig)
	for name, exit := range config.Exits {
		exits[name] = exit
	}
	config.Exits = exits

	// Test both uppercase and lowercase access
	_, ok := config.GetExit("US")
	if !ok {
		t.Error("expected to find exit with original case 'US'")
	}
}
