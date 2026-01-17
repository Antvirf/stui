package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	ArgumentOptions map[string]string `yaml:"argumentOptions"`
	Plugins         []PluginConfig    `yaml:"plugins"`
}

func LoadConfigsFromDirs(paths string) (Config, []string) {
	debugLogOutput := []string{}
	merged := NewConfig()
	for _, path := range strings.Split(paths, ",") {
		debugLogOutput = append(debugLogOutput, fmt.Sprintf("loading configs from directory: '%s'", path))
		files, err := os.ReadDir(path)
		if err != nil {
			debugLogOutput = append(debugLogOutput, fmt.Sprintf("failed to read config dir '%s': %v", path, err))
		}

		for _, file := range files {
			if filepath.Ext(file.Name()) != ".yaml" && filepath.Ext(file.Name()) != ".yml" {
				debugLogOutput = append(debugLogOutput, fmt.Sprintf("path: '%s' - skipping file '%s' due to bad extension (must be .yaml/.yml)", path, file.Name()))
				continue
			}
			debugLogOutput = append(debugLogOutput, fmt.Sprintf("path: '%s' - loading config from file: '%s'", path, file.Name()))

			cfg := loadConfig(filepath.Join(path, file.Name()))
			merged = mergeConfigs(merged, cfg)
		}
	}
	return merged, debugLogOutput
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
	newArgumentOptions := make(map[string]string)
	// Deep copy of old, then override with the new for all keys it defines
	for k, v := range base.ArgumentOptions {
		newArgumentOptions[k] = v
	}
	for k, v := range nextLayer.ArgumentOptions {
		newArgumentOptions[k] = v
	}

	merged := Config{
		ArgumentOptions: newArgumentOptions,
		Plugins:         append(base.Plugins, nextLayer.Plugins...),
	}
	return merged
}

func NewConfig() Config {
	return Config{
		ArgumentOptions: make(map[string]string),
		Plugins:         []PluginConfig{},
	}
}
