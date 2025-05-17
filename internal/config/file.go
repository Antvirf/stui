package config

import (
	"log"
	"os"

	"github.com/stretchr/testify/assert/yaml"
)

type PluginConfig struct {
	Name       string `yaml:"name"`
	ActivePage string `yaml:"activePage"`
	Shortcut   string `yaml:"shortcut"`
	Command    string `yaml:"command"`
}

type Config struct {
	Plugins []PluginConfig `yaml:"plugins"`
}

func LoadConfig(path string) Config {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file '%s': %v", path, err)
	}

	config := Config{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("failed to parse YAML from config file '%s': %v", path, err)
	}

	return config
}
