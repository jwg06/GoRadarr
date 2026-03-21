package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Host          string          `mapstructure:"host"`
	Port          int             `mapstructure:"port"`
	BaseURL       string          `mapstructure:"base_url"`
	LogLevel      string          `mapstructure:"log_level"`
	LogTarget     string          `mapstructure:"log_target"`    // stderr | stdout | file | syslog
	LogFile       string          `mapstructure:"log_file"`      // path when log_target=file
	SyslogAddress string          `mapstructure:"syslog_address"` // remote syslog host
	SyslogPort    int             `mapstructure:"syslog_port"`    // remote syslog port (default 514)
	SyslogNetwork string          `mapstructure:"syslog_network"` // udp | tcp | unix
	Database      DatabaseConfig  `mapstructure:"database"`
	Auth          AuthConfig      `mapstructure:"auth"`
	Data          DataConfig      `mapstructure:"data"`
	Metadata      MetadataConfig  `mapstructure:"metadata"`
	Scheduler     SchedulerConfig `mapstructure:"scheduler"`
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

// globalViper is retained so SaveToFile can persist runtime config changes.
var globalViper *viper.Viper

func Load() (*Config, error) {
	v := viper.New()
	globalViper = v

	v.SetDefault("host", "0.0.0.0")
	v.SetDefault("port", 7878)
	v.SetDefault("base_url", "")
	v.SetDefault("log_level", "info")
	v.SetDefault("log_target", "stderr")
	v.SetDefault("log_file", defaultDataDir("goradarr.log"))
	v.SetDefault("syslog_address", "")
	v.SetDefault("syslog_port", 514)
	v.SetDefault("syslog_network", "udp")
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

// SaveToFile persists the in-memory config back to config.yaml, so that
// runtime changes (e.g. syslog target) survive restarts.
func SaveToFile(cfg *Config) error {
	if globalViper == nil {
		return nil
	}
	globalViper.Set("host", cfg.Host)
	globalViper.Set("port", cfg.Port)
	globalViper.Set("base_url", cfg.BaseURL)
	globalViper.Set("log_level", cfg.LogLevel)
	globalViper.Set("log_target", cfg.LogTarget)
	globalViper.Set("log_file", cfg.LogFile)
	globalViper.Set("syslog_address", cfg.SyslogAddress)
	globalViper.Set("syslog_port", cfg.SyslogPort)
	globalViper.Set("syslog_network", cfg.SyslogNetwork)
	globalViper.Set("database.driver", cfg.Database.Driver)
	globalViper.Set("database.dsn", cfg.Database.DSN)
	globalViper.Set("auth.enabled", cfg.Auth.Enabled)
	globalViper.Set("auth.username", cfg.Auth.Username)
	globalViper.Set("auth.password_hash", cfg.Auth.PasswordHash)
	globalViper.Set("auth.api_key", cfg.Auth.APIKey)
	globalViper.Set("data.root_dir", cfg.Data.RootDir)
	globalViper.Set("metadata.tmdb_api_key", cfg.Metadata.TMDBAPIKey)
	globalViper.Set("scheduler.enabled", cfg.Scheduler.Enabled)
	globalViper.Set("scheduler.library_refresh_minutes", cfg.Scheduler.LibraryRefreshMinutes)
	globalViper.Set("scheduler.heartbeat_seconds", cfg.Scheduler.HeartbeatSeconds)
	return globalViper.WriteConfigAs(defaultDataDir("config.yaml"))
}
