package proxy

import (
	"testing"
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
