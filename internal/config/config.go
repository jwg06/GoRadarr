package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host     string         `mapstructure:"host"`
	Port     int            `mapstructure:"port"`
	BaseURL  string         `mapstructure:"base_url"`
	LogLevel string         `mapstructure:"log_level"`
	Database DatabaseConfig `mapstructure:"database"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Data     DataConfig     `mapstructure:"data"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"` // sqlite | postgres
	DSN    string `mapstructure:"dsn"`
}

type AuthConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	Username     string `mapstructure:"username"`
	PasswordHash string `mapstructure:"password_hash"`
	APIKey       string `mapstructure:"api_key"`
}

type DataConfig struct {
	RootDir string `mapstructure:"root_dir"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", 7878)
	v.SetDefault("base_url", "")
	v.SetDefault("log_level", "info")
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", defaultDataDir("goradarr.db"))
	v.SetDefault("auth.enabled", false)
	v.SetDefault("data.root_dir", defaultDataDir(""))

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(defaultDataDir(""))
	v.AddConfigPath(".")

	v.SetEnvPrefix("GORADARR")
	v.AutomaticEnv()

	_ = v.ReadInConfig()

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func defaultDataDir(rel string) string {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".config", "goradarr")
	_ = os.MkdirAll(base, 0755)
	if rel == "" {
		return base
	}
	return filepath.Join(base, rel)
}
