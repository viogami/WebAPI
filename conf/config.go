package conf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const defaultConfigPath = "conf/config.yaml"

type Config struct {
	Server ServerConfig `yaml:"server"`
	Text   TextConfig   `yaml:"text"`
	P5cc   P5ccConfig   `yaml:"p5cc"`
	AI     AIConfig     `yaml:"ai"`
	CH     CHAPIConfig  `yaml:"ch"`
}

type ServerConfig struct {
	Host           string          `yaml:"host"`
	Port           int             `yaml:"port"`
	GinMode        string          `yaml:"ginMode"`
	TrustedProxies []string        `yaml:"trustedProxies"`
	RateLimit      RateLimitConfig `yaml:"rateLimit"`
}

type RateLimitConfig struct {
	RequestsPerSecond float64 `yaml:"requestsPerSecond"`
	Burst             int     `yaml:"burst"`
	IdleTTLSeconds    int     `yaml:"idleTTLSeconds"`
}

type TextConfig struct {
	HelloText string `yaml:"helloText"`
}

type P5ccConfig struct {
	AssetPath  string  `yaml:"assetPath"`
	FontSize   float64 `yaml:"fontSize"`
	FontFamily string  `yaml:"fontFamily"`
	Gutter     float64 `yaml:"gutter"`
	Padding    float64 `yaml:"padding"`
	TextAlign  string  `yaml:"textAlign"`
	RedProb    float64 `yaml:"redProb"`
	ShowLogo   bool    `yaml:"showLogo"`
	LogoScale  float64 `yaml:"logoScale"`
	LogoOffset int     `yaml:"logoOffset"`
	ShowWtm    string  `yaml:"showWtm"`
}

type AIConfig struct {
	OpenAIBaseURL   string `yaml:"-"`
	DeepSeekBaseURL string `yaml:"-"`
	OpenAIAPIKey    string `yaml:"-"`
	DeepSeekAPIKey  string `yaml:"-"`
}

type CHAPIConfig struct {
	Enabled         bool   `yaml:"enabled"`
	SessionTTLHours int    `yaml:"sessionTTLHours"`
	DatabaseURL     string `yaml:"-"`
	PasswordPepper  string `yaml:"-"`
	AllowedOrigin   string `yaml:"-"`
}

func Load() (*Config, error) {
	return LoadFile(defaultConfigPath)
}

func LoadFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	applyDefaults(&cfg)
	if err := applyEnvironment(&cfg); err != nil {
		return nil, err
	}
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Server.RateLimit.RequestsPerSecond == 0 {
		cfg.Server.RateLimit.RequestsPerSecond = 5
	}
	if cfg.Server.RateLimit.Burst == 0 {
		cfg.Server.RateLimit.Burst = 20
	}
	if cfg.Server.RateLimit.IdleTTLSeconds == 0 {
		cfg.Server.RateLimit.IdleTTLSeconds = int((10 * time.Minute).Seconds())
	}
}

func applyEnvironment(cfg *Config) error {
	if port, ok := os.LookupEnv("PORT"); ok && strings.TrimSpace(port) != "" {
		value, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("parse PORT: %w", err)
		}
		cfg.Server.Port = value
	}

	cfg.AI.OpenAIBaseURL = strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
	cfg.AI.OpenAIAPIKey = strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	cfg.AI.DeepSeekBaseURL = strings.TrimSpace(os.Getenv("DEEPSEEK_BASE_URL"))
	cfg.AI.DeepSeekAPIKey = strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY"))
	cfg.CH.DatabaseURL = strings.TrimSpace(os.Getenv("CH_API_DATABASE_URL"))
	cfg.CH.PasswordPepper = strings.TrimSpace(os.Getenv("CH_API_PASSWORD_PEPPER"))
	cfg.CH.AllowedOrigin = getEnvOrDefault("CH_API_ALLOWED_ORIGIN", "http://localhost:3000")

	return nil
}

func validate(cfg Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}
	if cfg.Server.GinMode == "" {
		return fmt.Errorf("server ginMode is required")
	}
	if cfg.Server.RateLimit.RequestsPerSecond <= 0 {
		return fmt.Errorf("server rateLimit requestsPerSecond must be positive")
	}
	if cfg.Server.RateLimit.Burst <= 0 {
		return fmt.Errorf("server rateLimit burst must be positive")
	}
	if cfg.Server.RateLimit.IdleTTLSeconds <= 0 {
		return fmt.Errorf("server rateLimit idleTTLSeconds must be positive")
	}
	if cfg.P5cc.AssetPath == "" || cfg.P5cc.FontFamily == "" {
		return fmt.Errorf("p5cc assetPath and fontFamily are required")
	}
	if cfg.CH.SessionTTLHours <= 0 {
		return fmt.Errorf("ch sessionTTLHours must be positive")
	}
	if cfg.CH.Enabled {
		if cfg.CH.DatabaseURL == "" {
			return fmt.Errorf("CH_API_DATABASE_URL is required when ch is enabled")
		}
		if cfg.CH.PasswordPepper == "" {
			return fmt.Errorf("CH_API_PASSWORD_PEPPER is required when ch is enabled")
		}
	}
	return nil
}

func getEnvOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
