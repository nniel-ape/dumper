package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nerdneilsfield/dumper/internal/i18n"
	"github.com/nerdneilsfield/dumper/internal/ingest"
	"github.com/nerdneilsfield/dumper/internal/store"
)

// mediaGroupEntry buffers messages from a Telegram media group until the group is complete.
type mediaGroupEntry struct {
	messages []*tgbotapi.Message
	timer    *time.Timer
	mu       sync.Mutex
}

type Bot struct {
	api       *tgbotapi.BotAPI
	pipeline  *ingest.Pipeline
	stores    *store.Manager
	webAppURL string

	mediaGroups   map[string]*mediaGroupEntry
	mediaGroupsMu sync.Mutex
}

func New(token string, pipeline *ingest.Pipeline, stores *store.Manager, webAppURL string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("create bot api: %w", err)
	}

	slog.Info("authorized telegram bot", "username", api.Self.UserName)

	return &Bot{
		api:         api,
		pipeline:    pipeline,
		stores:      stores,
		webAppURL:   webAppURL,
		mediaGroups: make(map[string]*mediaGroupEntry),
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
			go func(u tgbotapi.Update) {
				defer func() {
					if r := recover(); r != nil {
						slog.Error("panic in update handler", "panic", r)
					}
				}()
				b.handleUpdate(ctx, u)
			}(update)
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
		"media_group_id", msg.MediaGroupID,
	)

	// Handle commands
	if msg.IsCommand() {
		b.handleCommand(ctx, msg)
		return
	}

	// Intercept media group photos
	if msg.MediaGroupID != "" && len(msg.Photo) > 0 {
		b.bufferMediaGroup(ctx, msg)
		return
	}

	// Handle regular messages
	b.handleMessage(ctx, msg)
}

const (
	// mediaGroupFlushDelay is the time to wait for more messages in a media group
	// before flushing. Telegram sends group messages within ~100-300ms of each other.
	mediaGroupFlushDelay = 500 * time.Millisecond
	// maxMediaGroupSize is the maximum number of photos Telegram allows in a media group.
	maxMediaGroupSize = 10
)

// bufferMediaGroup accumulates messages belonging to the same media group,
// then flushes them as a single batch after a short delay.
func (b *Bot) bufferMediaGroup(ctx context.Context, msg *tgbotapi.Message) {
	groupID := msg.MediaGroupID

	b.mediaGroupsMu.Lock()
	entry, exists := b.mediaGroups[groupID]
	if !exists {
		entry = &mediaGroupEntry{}
		b.mediaGroups[groupID] = entry
	}
	b.mediaGroupsMu.Unlock()

	entry.mu.Lock()
	entry.messages = append(entry.messages, msg)
	count := len(entry.messages)

	// Reset or start the flush timer
	if entry.timer != nil {
		entry.timer.Stop()
	}

	// Flush immediately if we've hit Telegram's max media group size
	if count >= maxMediaGroupSize {
		entry.mu.Unlock()
		b.flushMediaGroup(ctx, groupID)
		return
	}

	entry.timer = time.AfterFunc(mediaGroupFlushDelay, func() {
		if ctx.Err() != nil {
			// Context cancelled, clean up
			b.mediaGroupsMu.Lock()
			delete(b.mediaGroups, groupID)
			b.mediaGroupsMu.Unlock()
			return
		}
		b.flushMediaGroup(ctx, groupID)
	})
	entry.mu.Unlock()
}

// flushMediaGroup removes the group entry from the map and processes all buffered messages.
func (b *Bot) flushMediaGroup(ctx context.Context, groupID string) {
	b.mediaGroupsMu.Lock()
	entry, ok := b.mediaGroups[groupID]
	delete(b.mediaGroups, groupID)
	b.mediaGroupsMu.Unlock()

	if !ok {
		return
	}

	entry.mu.Lock()
	if entry.timer != nil {
		entry.timer.Stop()
	}
	messages := entry.messages
	entry.mu.Unlock()

	if len(messages) == 0 {
		return
	}

	// If only one photo arrived, handle as regular single photo
	if len(messages) == 1 {
		b.handlePhoto(ctx, messages[0])
		return
	}

	b.handleMediaGroup(ctx, messages)
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

func (b *Bot) sendAndReturn(chatID int64, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	return b.api.Send(msg)
}

func (b *Bot) edit(chatID int64, messageID int, text string) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "HTML"
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("failed to edit message", "error", err)
	}
}

func (b *Bot) editWithKeyboard(chatID int64, messageID int, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewEditMessageText(chatID, messageID, text)
	msg.ParseMode = "HTML"
	msg.ReplyMarkup = &keyboard
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("failed to edit message", "error", err)
	}
}

// getUserLang returns a Localizer for the user's preferred language.
// Priority: memory cache -> DB settings -> Telegram language code -> English default.
func (b *Bot) getUserLang(userID int64, telegramLangCode string) *i18n.Localizer {
	// 1. Check memory cache
	if lang, ok := i18n.GetCachedLang(userID); ok {
		return i18n.New(string(lang))
	}

	// 2. Check DB settings
	vault, err := b.stores.GetVault(userID)
	if err == nil {
		if langCode, err := vault.GetSetting("language"); err == nil {
			lang := i18n.ParseLang(langCode)
			i18n.CacheLang(userID, lang)
			return i18n.New(string(lang))
		} else if err != sql.ErrNoRows {
			slog.Warn("failed to get language setting", "user_id", userID, "error", err)
		}
	}

	// 3. Fall back to Telegram's LanguageCode
	lang := i18n.ParseLang(telegramLangCode)
	i18n.CacheLang(userID, lang)
	return i18n.New(string(lang))
}
