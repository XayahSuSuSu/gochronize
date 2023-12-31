package util

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Target struct {
	Url       string   `yaml:"url"`
	User      string   `yaml:"user"`
	Repo      string   `yaml:"repo"`
	Sync      string   `yaml:"sync"`
	Overwrite bool     `yaml:"overwrite"`
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

type HistoryAsset struct {
	Name               string `yaml:"name"`
	BrowserDownloadURL string `yaml:"browser_download_url"`
	CreatedAt          string `yaml:"created_at"`
	UpdatedAt          string `yaml:"updated_at"`
	ParentDir          string `yaml:"parent_dir"`
	FileName           string `yaml:"file_name"`
}

type HistoryRelease struct {
	Name    string         `yaml:"name"`
	TagName string         `yaml:"tag_name"`
	Assets  []HistoryAsset `yaml:"assets"`
}

type HistoryRepo struct {
	User     string           `yaml:"user"`
	Repo     string           `yaml:"repo"`
	Releases []HistoryRelease `yaml:"releases"`
}

type History struct {
	Repos []HistoryRepo `yaml:"repos"`
}

func ReadFromHistory(path string) *History {
	errMsg := "Failed to read from history"
	configFile, err := os.ReadFile(path)
	if err != nil {
		Fprintfln("%s: %v", errMsg, err)
	}

	history := History{}
	err = yaml.Unmarshal(configFile, &history)
	if err != nil {
		Fprintfln("%s: %v", errMsg, err)
	}

	return &history
}

func SaveHistoryToYaml(name string, history *History) error {
	data, err := yaml.Marshal(&history)
	if err != nil {
		return err
	}

	err = os.WriteFile(name, data, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}
