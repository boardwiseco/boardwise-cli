package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Token      string `json:"token,omitempty"`
	DefaultOrg string `json:"default_org,omitempty"`
	Superadmin bool   `json:"superadmin,omitempty"`
	URL        string `json:"url,omitempty"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "boardwise", "config.json"), nil
}

func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return &Config{}, nil
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) Clear() error {
	c.Token = ""
	c.DefaultOrg = ""
	c.Superadmin = false
	c.URL = ""
	return c.Save()
}
