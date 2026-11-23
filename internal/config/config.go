// Package config loads local operator configuration.
package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config contains the local RPC endpoints and credentials required by operator commands.
type Config struct {
	RPC     map[string][]string `yaml:"rpc"`
	Signing struct {
		PrivateKey string `yaml:"privateKey"`
	} `yaml:"signing"`
	Explorer struct {
		APIKey string `yaml:"apiKey"`
	} `yaml:"explorer"`
}

// Load reads and validates one local YAML configuration file.
func Load(path string) (*Config, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errors.New("configuration file not found")
		}
		return nil, errors.New("open configuration")
	}

	configuration := &Config{}
	if err := yaml.Unmarshal(contents, configuration); err != nil {
		return nil, errors.New("parse configuration")
	}
	if err := configuration.Validate(); err != nil {
		return nil, err
	}
	return configuration, nil
}

// Validate checks the complete configuration required before command execution.
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("configuration is required")
	}
	if c.RPC == nil {
		return errors.New("rpc is required")
	}
	for _, network := range []string{"eth", "arb"} {
		endpoints := c.RPC[network]
		if len(endpoints) == 0 {
			return fmt.Errorf("rpc.%s is required", network)
		}
		for _, endpoint := range endpoints {
			if strings.TrimSpace(endpoint) == "" {
				return fmt.Errorf("rpc.%s contains a blank endpoint", network)
			}
		}
	}
	if strings.TrimSpace(c.Signing.PrivateKey) == "" {
		return errors.New("signing.privateKey is required")
	}
	if strings.TrimSpace(c.Explorer.APIKey) == "" {
		return errors.New("explorer.apiKey is required")
	}
	return nil
}
