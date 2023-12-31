package util

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Target struct {
	User      string   `yaml:"user"`
	Repo      string   `yaml:"repo"`
	Sync      string   `yaml:"sync"`
	ParentDir string   `yaml:"parent_dir"`
	FileName  string   `yaml:"file_name"`
	Exclusion []string `yaml:"exclusion"`
}

type Config struct {
	ProxyHttp  string `yaml:"proxy_http"`
	Timeout    int    `yaml:"timeout"`
	Retries    int    `yaml:"retries"`
	TimeFormat string `yaml:"time_format"`

	Targets []Target `yaml:"targets"`
}

func ReadFromConfig(path string) *Config {
	errMsg := "Failed to read from config"
	configFile, err := os.ReadFile(path)
	if err != nil {
		Fprintfln("%s: %v", errMsg, err)
		os.Exit(ErrorIO)
	}

	config := Config{}
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		Fprintfln("%s: %v", errMsg, err)
		os.Exit(ErrorIO)
	}

	return &config
}
