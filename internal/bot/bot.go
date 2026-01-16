package bot

import (
	"context"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nerdneilsfield/dumper/internal/ingest"
	"github.com/nerdneilsfield/dumper/internal/store"
)

type Bot struct {
	api       *tgbotapi.BotAPI
	pipeline  *ingest.Pipeline
	stores    *store.Manager
	webAppURL string
}

func New(token string, pipeline *ingest.Pipeline, stores *store.Manager, webAppURL string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot api: %w", err)
	}

	slog.Info("authorized telegram bot", "username", api.Self.UserName)

	return &Bot{
		api:       api,
		pipeline:  pipeline,
		stores:    stores,
		webAppURL: webAppURL,
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case update := <-updates:
			go b.handleUpdate(ctx, update)
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	userID := msg.From.ID

	slog.Debug("received message",
		"user_id", userID,
		"text", msg.Text,
		"has_entities", len(msg.Entities) > 0,
	)

	// Handle commands
	if msg.IsCommand() {
		b.handleCommand(ctx, msg)
		return
	}

	// Handle regular messages
	b.handleMessage(ctx, msg)
}

func (b *Bot) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("failed to send message", "error", err)
	}
}

func (b *Bot) sendWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("failed to send message", "error", err)
	}
}
