package config

import (
	"github.com/jessevdk/go-flags"
)

type Config struct {
	TelegramToken   string `long:"telegram-token" env:"TELEGRAM_BOT_TOKEN" description:"Telegram bot token" required:"true"`
	OpenRouterKey   string `long:"openrouter-key" env:"OPENROUTER_API_KEY" description:"OpenRouter API key" required:"true"`
	DataDir         string `long:"data-dir" env:"DATA_DIR" default:"./data" description:"Data directory for SQLite databases"`
	HTTPPort        int    `long:"http-port" env:"HTTP_PORT" default:"8080" description:"HTTP server port"`
	LogLevel        string `long:"log-level" env:"LOG_LEVEL" default:"info" description:"Log level: debug|info|warn|error"`
	OpenRouterModel string `long:"openrouter-model" env:"OPENROUTER_MODEL" default:"anthropic/claude-3-haiku" description:"OpenRouter model ID"`
	WebAppURL       string `long:"webapp-url" env:"WEBAPP_URL" description:"Telegram Mini App URL"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	parser := flags.NewParser(cfg, flags.Default)
	if _, err := parser.Parse(); err != nil {
		return nil, err
	}
	return cfg, nil
}
