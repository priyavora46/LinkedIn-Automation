package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Browser  BrowserConfig  `yaml:"browser"`
	LinkedIn LinkedInConfig `yaml:"linkedin"`
	Limits   LimitsConfig   `yaml:"limits"`
	Delays   DelaysConfig   `yaml:"delays"`
	Stealth  StealthConfig  `yaml:"stealth"`
	Storage  StorageConfig  `yaml:"storage"`
	Logging  LoggingConfig  `yaml:"logging"`
	Creds    CredsConfig
}

type BrowserConfig struct {
	Headless  bool   `yaml:"headless"`
	Width     int    `yaml:"width"`
	Height    int    `yaml:"height"`
	UserAgent string `yaml:"user_agent"`
}

type LinkedInConfig struct {
	BaseURL   string `yaml:"base_url"`
	LoginURL  string `yaml:"login_url"`
	SearchURL string `yaml:"search_url"`
}

type LimitsConfig struct {
	MaxConnectionsPerDay int `yaml:"max_connections_per_day"`
	MaxMessagesPerDay    int `yaml:"max_messages_per_day"`
	ConnectionNoteMaxLen int `yaml:"connection_note_max_length"`
}

type DelaysConfig struct {
	MinActionDelayMs int `yaml:"min_action_delay_ms"`
	MaxActionDelayMs int `yaml:"max_action_delay_ms"`
	MinTypingDelayMs int `yaml:"min_typing_delay_ms"`
	MaxTypingDelayMs int `yaml:"max_typing_delay_ms"`
	MinScrollDelayMs int `yaml:"min_scroll_delay_ms"`
	MaxScrollDelayMs int `yaml:"max_scroll_delay_ms"`
}

type StealthConfig struct {
	BusinessHoursOnly     bool    `yaml:"business_hours_only"`
	WorkStartHour         int     `yaml:"work_start_hour"`
	WorkEndHour           int     `yaml:"work_end_hour"`
	EnableRandomScrolling bool    `yaml:"enable_random_scrolling"`
	EnableMouseHovering   bool    `yaml:"enable_mouse_hovering"`
	EnableTypingErrors    bool    `yaml:"enable_typing_errors"`
	TypoProbability       float64 `yaml:"typo_probability"`
}

type StorageConfig struct {
	DBPath            string `yaml:"db_path"`
	SessionCookiePath string `yaml:"session_cookie_path"`
}

type LoggingConfig struct {
	Level   string `yaml:"level"`
	File    string `yaml:"file"`
	Console bool   `yaml:"console"`
}

type CredsConfig struct {
	Email    string
	Password string
}

func Load(configPath string) (*Config, error) {
	// Load .env file
	_ = godotenv.Load()

	// Read YAML config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Load credentials from environment
	cfg.Creds.Email = os.Getenv("LINKEDIN_EMAIL")
	cfg.Creds.Password = os.Getenv("LINKEDIN_PASSWORD")

	// Override with environment variables if present
	if val := os.Getenv("HEADLESS"); val != "" {
		cfg.Browser.Headless, _ = strconv.ParseBool(val)
	}
	if val := os.Getenv("MAX_CONNECTIONS_PER_DAY"); val != "" {
		cfg.Limits.MaxConnectionsPerDay, _ = strconv.Atoi(val)
	}
	if val := os.Getenv("LOG_LEVEL"); val != "" {
		cfg.Logging.Level = val
	}

	// Validate required fields
	if cfg.Creds.Email == "" || cfg.Creds.Password == "" {
		return nil, fmt.Errorf("LINKEDIN_EMAIL and LINKEDIN_PASSWORD must be set")
	}

	return cfg, nil
}
