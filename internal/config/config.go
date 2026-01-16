package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	TelegramToken   string `env:"TELEGRAM_BOT_TOKEN,required"`
	OpenRouterKey   string `env:"OPENROUTER_API_KEY,required"`
	DataDir         string `env:"DATA_DIR" envDefault:"./data"`
	HTTPPort        int    `env:"HTTP_PORT" envDefault:"8080"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
	OpenRouterModel string `env:"OPENROUTER_MODEL" envDefault:"anthropic/claude-3-haiku"`
	WebAppURL       string `env:"WEBAPP_URL" envDefault:""`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
