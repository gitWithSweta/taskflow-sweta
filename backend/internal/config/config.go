package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	App    AppConfig    `yaml:"app"`
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
	Auth   AuthConfig   `yaml:"auth"`
	CORS   CORSConfig   `yaml:"cors"`
	Seed   SeedConfig   `yaml:"seed"`
}

type AppConfig struct {
	Name     string   `yaml:"name"`
	Env      string   `yaml:"env"`
	LogLevel LogLevel `yaml:"log_level"`
}

type ServerConfig struct {
	Port            int      `yaml:"port"`
	ReadTimeout     Duration `yaml:"read_timeout"`
	WriteTimeout    Duration `yaml:"write_timeout"`
	IdleTimeout     Duration `yaml:"idle_timeout"`
	ShutdownTimeout Duration `yaml:"shutdown_timeout"`
}

type DBConfig struct {
	URL  string     `yaml:"url"`
	Pool PoolConfig `yaml:"pool"`
}

type PoolConfig struct {
	MaxConns              int32    `yaml:"max_conns"`
	MinConns              int32    `yaml:"min_conns"`
	MaxConnLifetime       Duration `yaml:"max_conn_lifetime"`
	MaxConnLifetimeJitter Duration `yaml:"max_conn_lifetime_jitter"`
	MaxConnIdleTime       Duration `yaml:"max_conn_idle_time"`
	HealthCheckPeriod     Duration `yaml:"health_check_period"`
	ConnectTimeout        Duration `yaml:"connect_timeout"`
	StatementTimeout      Duration `yaml:"statement_timeout"`
}

type AuthConfig struct {
	JWTSecret string   `yaml:"jwt_secret"`
	TokenTTL  Duration `yaml:"token_ttl"`
}

type CORSConfig struct {
	AllowedOrigins string `yaml:"allowed_origins"`
}

type SeedConfig struct {
	CSVDir string `yaml:"csv_dir"`
}

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf(":%d", c.Server.Port)
}

func (c *Config) CORSOrigins() []string {
	if c.CORS.AllowedOrigins == "" {
		return nil
	}
	var out []string
	for _, o := range strings.Split(c.CORS.AllowedOrigins, ",") {
		if t := strings.TrimSpace(o); t != "" {
			out = append(out, t)
		}
	}
	return out
}

type Duration struct{ time.Duration }

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	dur, err := time.ParseDuration(value.Value)
	if err != nil {
		return fmt.Errorf("config: invalid duration %q: %w", value.Value, err)
	}
	d.Duration = dur
	return nil
}

type LogLevel struct{ slog.Level }

func (l *LogLevel) UnmarshalYAML(value *yaml.Node) error {
	if err := l.Level.UnmarshalText([]byte(strings.ToUpper(value.Value))); err != nil {
		return fmt.Errorf("config: invalid log_level %q (valid: DEBUG INFO WARN ERROR): %w", value.Value, err)
	}
	return nil
}

func MustLoad() *Config {

	_ = godotenv.Load("../.env")
	_ = godotenv.Load()
	cfg, err := Load()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}
	return cfg
}

func Load() (*Config, error) {
	profile := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if profile == "" {
		profile = "dev"
	}

	raw, err := readConfigFile(profile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse application-%s.yml: %w", profile, err)
	}

	cfg.applyEnvOverrides()

	cfg.applyDefaults()

	return &cfg, cfg.validate()
}

func (c *Config) applyEnvOverrides() {

	if v := os.Getenv("APP_LOG_LEVEL"); v != "" {
		var lvl slog.Level
		if err := lvl.UnmarshalText([]byte(strings.ToUpper(v))); err == nil {
			c.App.LogLevel.Level = lvl
		}
	}

	if v := os.Getenv("HTTP_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil && port > 0 {
			c.Server.Port = port
		}
	}

	if v := os.Getenv("DATABASE_URL"); v != "" {
		c.DB.URL = v
	}

	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.Auth.JWTSecret = v
	}

	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		c.CORS.AllowedOrigins = v
	}

	if v := os.Getenv("SEED_CSV_DIR"); v != "" {
		c.Seed.CSVDir = v
	}
}

func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 4000
	}
	if c.Server.ReadTimeout.Duration == 0 {
		c.Server.ReadTimeout.Duration = 15 * time.Second
	}
	if c.Server.WriteTimeout.Duration == 0 {
		c.Server.WriteTimeout.Duration = 30 * time.Second
	}
	if c.Server.IdleTimeout.Duration == 0 {
		c.Server.IdleTimeout.Duration = 60 * time.Second
	}
	if c.Server.ShutdownTimeout.Duration == 0 {
		c.Server.ShutdownTimeout.Duration = 15 * time.Second
	}
	if c.DB.Pool.MaxConns == 0 {
		c.DB.Pool.MaxConns = 10
	}
	if c.DB.Pool.MinConns == 0 {
		c.DB.Pool.MinConns = 2
	}
	if c.DB.Pool.MaxConnLifetime.Duration == 0 {
		c.DB.Pool.MaxConnLifetime.Duration = 1 * time.Hour
	}
	if c.DB.Pool.MaxConnIdleTime.Duration == 0 {
		c.DB.Pool.MaxConnIdleTime.Duration = 30 * time.Minute
	}
	if c.DB.Pool.HealthCheckPeriod.Duration == 0 {
		c.DB.Pool.HealthCheckPeriod.Duration = time.Minute
	}
	if c.DB.Pool.ConnectTimeout.Duration == 0 {
		c.DB.Pool.ConnectTimeout.Duration = 10 * time.Second
	}
	if c.Auth.TokenTTL.Duration == 0 {
		c.Auth.TokenTTL.Duration = 24 * time.Hour
	}
}

func (c *Config) validate() error {
	if c.DB.URL == "" {
		return fmt.Errorf("db.url is required — set DATABASE_URL env var or db.url in the config file")
	}
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("auth.jwt_secret is required — set JWT_SECRET env var or auth.jwt_secret in the config file")
	}
	return nil
}

func readConfigFile(profile string) ([]byte, error) {
	filename := fmt.Sprintf("application-%s.yml", profile)

	if path := strings.TrimSpace(os.Getenv("CONFIG_PATH")); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("CONFIG_PATH=%s: %w", path, err)
		}
		return b, nil
	}

	if b, err := os.ReadFile(filepath.Join("config", filename)); err == nil {
		return b, nil
	}

	if exe, err := os.Executable(); err == nil {
		if b, err := os.ReadFile(filepath.Join(filepath.Dir(exe), "config", filename)); err == nil {
			return b, nil
		}
	}

	return nil, fmt.Errorf(
		"config file %q not found in ./config/ or alongside the executable\n"+
			"  → APP_ENV=%q  (change with APP_ENV=prd)\n"+
			"  → or set CONFIG_PATH to an explicit file path",
		filename, profile,
	)
}
