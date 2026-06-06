// Package config persists the API host and bearer token between invocations.
// Tokens go to the OS keychain when available, falling back to a 0600 file.
package config

import (
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
	"gopkg.in/yaml.v3"
)

const keyringService = "superdocu-cli"

// Config is the on-disk configuration file.
type Config struct {
	Host  string `yaml:"host,omitempty"`
	Token string `yaml:"token,omitempty"` // only used when the keychain is unavailable
}

// Dir is the configuration directory (~/.config/superdocu on Linux/macOS).
func Dir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "superdocu"), nil
}

func filePath() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.yaml"), nil
}

// Path returns the config file path for display purposes.
func Path() string {
	p, _ := filePath()
	return p
}

// Load reads the config file, returning an empty config when absent.
func Load() (*Config, error) {
	p, err := filePath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Save writes the config file with restrictive permissions.
func (c *Config) Save() error {
	d, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(d, 0o700); err != nil {
		return err
	}
	p, _ := filePath()
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o600)
}

// SaveToken stores the token in the keychain, falling back to the config file.
func SaveToken(host, token string) error {
	if err := keyring.Set(keyringService, host, token); err == nil {
		if c, _ := Load(); c != nil && c.Token != "" {
			c.Token = ""
			_ = c.Save()
		}
		return nil
	}
	c, err := Load()
	if err != nil {
		return err
	}
	c.Token = token
	return c.Save()
}

// LoadToken returns the stored token for host, preferring the keychain.
func LoadToken(host string) string {
	if t, err := keyring.Get(keyringService, host); err == nil && t != "" {
		return t
	}
	if c, _ := Load(); c != nil {
		return c.Token
	}
	return ""
}

// DeleteToken removes the stored token from both keychain and config file.
func DeleteToken(host string) {
	_ = keyring.Delete(keyringService, host)
	if c, _ := Load(); c != nil && c.Token != "" {
		c.Token = ""
		_ = c.Save()
	}
}
