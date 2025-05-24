package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert/yaml"
)

type PluginConfig struct {
	Name                    string `yaml:"name"`
	ActivePage              string `yaml:"activePage"`
	Shortcut                string `yaml:"shortcut"`
	Command                 string `yaml:"command"`
	ExecuteImmediately      bool   `yaml:"executeImmediately"`
	ClosePromptAfterExecute bool   `yaml:"closePromptAfterExecute"`
}

type Config struct {
	Plugins []PluginConfig `yaml:"plugins"`
}

func LoadConfigsFromDir(path string) Config {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("failed to read config dir '%s': %v", path, err)
	}

	merged := NewConfig()
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml" {
			continue
		}

		cfg := loadConfig(filepath.Join(path, file.Name()))
		merged = mergeConfigs(merged, cfg)
	}
	return merged
}

func loadConfig(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file '%s': %v", path, err)
	}

	config := NewConfig()
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("failed to parse YAML from config file '%s': %v", path, err)
	}

	return config
}

// Merges two configs, with the nextLayer config taking precedence on specific keys. Arrays
// are concatenated, and maps are merged.
// This is a custom implementation and needs updating as the config structure changes.
func mergeConfigs(base Config, nextLayer Config) Config {
	merged := Config{
		Plugins: append(base.Plugins, nextLayer.Plugins...),
	}
	return merged
}

func NewConfig() Config {
	return Config{
		Plugins: []PluginConfig{},
	}
}
