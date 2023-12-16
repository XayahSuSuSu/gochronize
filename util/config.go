package util

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Target struct {
	User string `yaml:"user"`
	Repo string `yaml:"repo"`
}

type Config struct {
	ProxyHttp string `yaml:"proxy_http"`

	Targets []Target `yaml:"targets"`
}

func ReadFromConfig(path string) (*Config, error) {
	configFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := Config{}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
