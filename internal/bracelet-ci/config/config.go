package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is persisted to ~/.bracelet-ci/config.yaml on `bracelet-ci init`.
type Config struct {
	WorkerToken      string `yaml:"worker_token"`
	RedisPassword    string `yaml:"redis_password"`
	RedisURL         string `yaml:"redis_url"`
	PostgresPassword string `yaml:"postgres_password"`
	PostgresDSN      string `yaml:"postgres_dsn"`
}

// Dir returns ~/.bracelet-ci
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".bracelet-ci"), nil
}

// Path returns the full path to the config file.
func Path() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.yaml"), nil
}

// EnvFilePath returns the path to ~/.bracelet-ci/.env
func EnvFilePath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, ".env"), nil
}

// ComposeFilePath returns the path to ~/.bracelet-ci/docker-compose.yml
func ComposeFilePath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "docker-compose.yml"), nil
}

// Save writes cfg to ~/.bracelet-ci/config.yaml, creating the directory if needed.
func Save(cfg Config) error {
	d, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0700); err != nil {
		return err
	}
	path := filepath.Join(d, "config.yaml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	// 0600 — readable only by the owner; file contains secrets
	return os.WriteFile(path, data, 0600)
}

// Load reads ~/.bracelet-ci/config.yaml into a Config.
func Load() (Config, error) {
	path, err := Path()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, errors.New("no config found — run `bracelet-ci init` first")
		}
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
