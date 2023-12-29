package util

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type Target struct {
	User string `yaml:"user"`
	Repo string `yaml:"repo"`
	Sync string `yaml:"sync"`
}

type Config struct {
	ProxyHttp string `yaml:"proxy_http"`
	Timeout   int    `yaml:"timeout"`
	Retries   int    `yaml:"retries"`

	Targets []Target `yaml:"targets"`
}

func ReadFromConfig(path string) (*Config, error) {
	errMsg := "Failed to read from config"
	configFile, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("%s: %s.\n", errMsg, err.Error())
		return nil, err
	}

	config := Config{}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Printf("%s: %s.\n", errMsg, err.Error())
		return nil, err
	}

	return &config, nil
}
