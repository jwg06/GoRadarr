package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host      string          `mapstructure:"host"`
	Port      int             `mapstructure:"port"`
	BaseURL   string          `mapstructure:"base_url"`
	LogLevel  string          `mapstructure:"log_level"`
	LogTarget string          `mapstructure:"log_target"` // stderr | stdout | file | syslog
	LogFile   string          `mapstructure:"log_file"`   // path when log_target=file
	Database  DatabaseConfig  `mapstructure:"database"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Data      DataConfig      `mapstructure:"data"`
	Metadata  MetadataConfig  `mapstructure:"metadata"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
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

type MetadataConfig struct {
	TMDBAPIKey string `mapstructure:"tmdb_api_key"`
}

type SchedulerConfig struct {
	Enabled               bool `mapstructure:"enabled"`
	LibraryRefreshMinutes int  `mapstructure:"library_refresh_minutes"`
	HeartbeatSeconds      int  `mapstructure:"heartbeat_seconds"`
}

func Load() (*Config, error) {
	v := viper.New()

	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", 7878)
	v.SetDefault("base_url", "")
	v.SetDefault("log_level", "info")
	v.SetDefault("log_target", "stderr")
	v.SetDefault("log_file", defaultDataDir("goradarr.log"))
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", defaultDataDir("goradarr.db"))
	v.SetDefault("auth.enabled", false)
	v.SetDefault("data.root_dir", defaultDataDir(""))
	v.SetDefault("metadata.tmdb_api_key", "")
	v.SetDefault("scheduler.enabled", true)
	v.SetDefault("scheduler.library_refresh_minutes", 15)
	v.SetDefault("scheduler.heartbeat_seconds", 30)

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
	_ = os.MkdirAll(base, 0o755)
	if rel == "" {
		return base
	}
	return filepath.Join(base, rel)
}
