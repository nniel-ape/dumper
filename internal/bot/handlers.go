package bot

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/nerdneilsfield/dumper/internal/i18n"
	"github.com/nerdneilsfield/dumper/internal/ingest"
)

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message) {
	switch msg.Command() {
	case "start":
		b.handleStart(msg)
	case "help":
		b.handleHelp(msg)
	case "search":
		b.handleSearch(ctx, msg)
	case "recent":
		b.handleRecent(ctx, msg)
	case "tags":
		b.handleTags(ctx, msg)
	case "export":
		b.handleExport(ctx, msg)
	case "app":
		b.handleApp(msg)
	case "stats":
		b.handleStats(ctx, msg)
	case "lang":
		b.handleLang(ctx, msg)
	default:
		l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)
		b.send(msg.Chat.ID, l.Get(i18n.MsgUnknownCommand))
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)
	text := l.Get(i18n.MsgWelcome)

	if b.webAppURL != "" {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(l.Get(i18n.MsgOpenApp), b.webAppURL),
			),
		)
		b.sendWithKeyboard(msg.Chat.ID, text, keyboard)
	} else {
		b.send(msg.Chat.ID, text)
	}
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)
	b.send(msg.Chat.ID, l.Get(i18n.MsgHelp))
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	// Check for photo FIRST
	if len(msg.Photo) > 0 {
		b.handlePhoto(ctx, msg)
		return
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return
	}

	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	var raw ingest.RawContent
	raw.UserID = msg.From.ID
	raw.Language = l.Code()

	var sentMsg tgbotapi.Message
	if ingest.IsURL(text) {
		sentMsg, _ = b.sendAndReturn(msg.Chat.ID, l.Get(i18n.MsgProcessingLink))
		raw.Type = ingest.ContentTypeLink
		raw.URL = text
	} else if ingest.IsShortTopicMessage(text) {
		sentMsg, _ = b.sendAndReturn(msg.Chat.ID, l.Getf(i18n.MsgSearching, text))
		raw.Type = ingest.ContentTypeSearch
		raw.Text = text
	} else {
		sentMsg, _ = b.sendAndReturn(msg.Chat.ID, l.Get(i18n.MsgProcessingNote))
		raw.Type = ingest.ContentTypeNote
		raw.Text = text
	}

	item, err := b.pipeline.Process(ctx, raw)
	if err != nil {
		b.edit(msg.Chat.ID, sentMsg.MessageID, l.Getf(i18n.MsgFailedProcess, err))
		return
	}

	// Format response
	var tagsStr string
	if len(item.Tags) > 0 {
		tagsStr = "#" + strings.Join(item.Tags, " #")
	}

	response := fmt.Sprintf(`%s

<b>%s</b>

%s

%s`, l.Get(i18n.MsgSaved), item.Title, item.Summary, tagsStr)

	if b.webAppURL != "" {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(l.Get(i18n.MsgViewInApp), b.webAppURL+"?item="+item.ID),
			),
		)
		b.editWithKeyboard(msg.Chat.ID, sentMsg.MessageID, response, keyboard)
	} else {
		b.edit(msg.Chat.ID, sentMsg.MessageID, response)
	}
}

// downloadPhoto downloads the largest resolution of a photo from a Telegram message.
func (b *Bot) downloadPhoto(ctx context.Context, msg *tgbotapi.Message) ([]byte, string, error) {
	photos := msg.Photo
	photo := photos[len(photos)-1]

	fileConfig := tgbotapi.FileConfig{FileID: photo.FileID}
	file, err := b.api.GetFile(fileConfig)
	if err != nil {
		return nil, "", fmt.Errorf("get file info: %w", err)
	}

	fileURL := file.Link(b.api.Token)

	if !strings.HasPrefix(fileURL, "https://api.telegram.org/file/") {
		return nil, "", fmt.Errorf("invalid telegram file URL")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}()

	const maxImageSize = 20 * 1024 * 1024
	limitedReader := io.LimitReader(resp.Body, maxImageSize+1)
	imageData, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, "", fmt.Errorf("read image: %w", err)
	}

	if len(imageData) > maxImageSize {
		return nil, "", fmt.Errorf("image too large (max 20MB)")
	}

	ext := "jpg"
	if filePath := file.FilePath; filePath != "" {
		if e := path.Ext(filePath); e != "" {
			ext = strings.TrimPrefix(e, ".")
		}
	}

	return imageData, ext, nil
}

func (b *Bot) handlePhoto(ctx context.Context, msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	sentMsg, _ := b.sendAndReturn(msg.Chat.ID, l.Get(i18n.MsgSavingImage))

	imageData, ext, err := b.downloadPhoto(ctx, msg)
	if err != nil {
		b.edit(msg.Chat.ID, sentMsg.MessageID, l.Getf(i18n.MsgFailedDownload, err))
		return
	}

	raw := ingest.RawContent{
		Type:      ingest.ContentTypeImage,
		UserID:    msg.From.ID,
		ImageData: imageData,
		ImageExt:  ext,
		Caption:   msg.Caption,
		Language:  l.Code(),
	}

	item, err := b.pipeline.Process(ctx, raw)
	if err != nil {
		b.edit(msg.Chat.ID, sentMsg.MessageID, l.Getf(i18n.MsgFailedSaveImage, err))
		return
	}

	var tagsStr string
	if len(item.Tags) > 0 {
		tagsStr = "#" + strings.Join(item.Tags, " #")
	}

	response := fmt.Sprintf(`%s

<b>%s</b>

%s`, l.Get(i18n.MsgImageSaved), item.Title, tagsStr)

	if b.webAppURL != "" {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(l.Get(i18n.MsgViewInApp), b.webAppURL+"?item="+item.ID),
			),
		)
		b.editWithKeyboard(msg.Chat.ID, sentMsg.MessageID, response, keyboard)
	} else {
		b.edit(msg.Chat.ID, sentMsg.MessageID, response)
	}
}

// handleMediaGroup processes a batch of messages that form a Telegram media group.
func (b *Bot) handleMediaGroup(ctx context.Context, messages []*tgbotapi.Message) {
	first := messages[0]
	l := b.getUserLang(first.From.ID, first.From.LanguageCode)

	sentMsg, _ := b.sendAndReturn(first.Chat.ID, l.Getf(i18n.MsgSavingImages, len(messages)))

	// Find caption from any message in the group (usually the first)
	var caption string
	for _, msg := range messages {
		if msg.Caption != "" {
			caption = msg.Caption
			break
		}
	}

	// Download all photos in parallel
	type downloadResult struct {
		index int
		data  []byte
		ext   string
		err   error
	}

	results := make([]downloadResult, len(messages))
	var wg sync.WaitGroup
	for i, msg := range messages {
		wg.Add(1)
		go func(idx int, m *tgbotapi.Message) {
			defer wg.Done()
			data, ext, err := b.downloadPhoto(ctx, m)
			results[idx] = downloadResult{index: idx, data: data, ext: ext, err: err}
		}(i, msg)
	}
	wg.Wait()

	// Collect successful downloads (preserve order)
	var images []ingest.ImageFile
	for _, r := range results {
		if r.err != nil {
			slog.Warn("failed to download media group photo",
				"index", r.index, "error", r.err)
			continue
		}
		images = append(images, ingest.ImageFile{Data: r.data, Ext: r.ext})
	}

	if len(images) == 0 {
		b.edit(first.Chat.ID, sentMsg.MessageID, l.Getf(i18n.MsgFailedDownload, fmt.Errorf("all downloads failed")))
		return
	}

	raw := ingest.RawContent{
		Type:     ingest.ContentTypeImage,
		UserID:   first.From.ID,
		Images:   images,
		Caption:  caption,
		Language: l.Code(),
	}

	item, err := b.pipeline.Process(ctx, raw)
	if err != nil {
		b.edit(first.Chat.ID, sentMsg.MessageID, l.Getf(i18n.MsgFailedSaveImage, err))
		return
	}

	var tagsStr string
	if len(item.Tags) > 0 {
		tagsStr = "#" + strings.Join(item.Tags, " #")
	}

	savedMsg := l.Getf(i18n.MsgImagesSaved, len(images))
	response := fmt.Sprintf(`%s

<b>%s</b>

%s`, savedMsg, item.Title, tagsStr)

	if b.webAppURL != "" {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL(l.Get(i18n.MsgViewInApp), b.webAppURL+"?item="+item.ID),
			),
		)
		b.editWithKeyboard(first.Chat.ID, sentMsg.MessageID, response, keyboard)
	} else {
		b.edit(first.Chat.ID, sentMsg.MessageID, response)
	}
}

func (b *Bot) handleSearch(ctx context.Context, msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	query := msg.CommandArguments()
	if query == "" {
		b.send(msg.Chat.ID, l.Get(i18n.MsgSearchUsage))
		return
	}

	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, l.Get(i18n.MsgFailedVault))
		return
	}

	results, err := vault.Search(query, 5)
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedSearch, err))
		return
	}

	if len(results) == 0 {
		b.send(msg.Chat.ID, l.Get(i18n.MsgNoResults))
		return
	}

	var text strings.Builder
	text.WriteString(l.Getf(i18n.MsgSearchFor, query))
	text.WriteString("\n\n")

	for i, r := range results {
		text.WriteString(fmt.Sprintf("%d. <b>%s</b>\n", i+1, r.Item.Title))
		if r.Snippet != "" {
			text.WriteString(fmt.Sprintf("   %s\n", r.Snippet))
		}
		text.WriteString("\n")
	}

	b.send(msg.Chat.ID, text.String())
}

func (b *Bot) handleRecent(ctx context.Context, msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, l.Get(i18n.MsgFailedVault))
		return
	}

	items, err := vault.ListItems(5, 0)
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedListItems, err))
		return
	}

	if len(items) == 0 {
		b.send(msg.Chat.ID, l.Get(i18n.MsgNoItems))
		return
	}

	var text strings.Builder
	text.WriteString(l.Get(i18n.MsgRecentItems))

	for i, item := range items {
		text.WriteString(fmt.Sprintf("%d. <b>%s</b>\n", i+1, item.Title))
		if len(item.Tags) > 0 {
			text.WriteString(fmt.Sprintf("   #%s\n", strings.Join(item.Tags, " #")))
		}
		text.WriteString("\n")
	}

	b.send(msg.Chat.ID, text.String())
}

func (b *Bot) handleTags(ctx context.Context, msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, l.Get(i18n.MsgFailedVault))
		return
	}

	tags, err := vault.GetAllTags()
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedGetTags, err))
		return
	}

	if len(tags) == 0 {
		b.send(msg.Chat.ID, l.Get(i18n.MsgNoTags))
		return
	}

	b.send(msg.Chat.ID, l.Getf(i18n.MsgYourTags, strings.Join(tags, " #")))
}

func (b *Bot) handleStats(ctx context.Context, msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, l.Get(i18n.MsgFailedVault))
		return
	}

	count, err := vault.ItemCount()
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedGetStats, err))
		return
	}

	tags, err := vault.GetAllTags()
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedGetStats, err))
		return
	}

	b.send(msg.Chat.ID, l.Getf(i18n.MsgYourVault, count, len(tags)))
}

func (b *Bot) handleExport(ctx context.Context, msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)
	b.send(msg.Chat.ID, l.Get(i18n.MsgExportComingSoon))
}

func (b *Bot) handleApp(msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	if b.webAppURL == "" {
		b.send(msg.Chat.ID, l.Get(i18n.MsgAppNotConfigured))
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(l.Get(i18n.MsgOpenApp), b.webAppURL),
		),
	)
	b.sendWithKeyboard(msg.Chat.ID, l.Get(i18n.MsgOpenMiniApp), keyboard)
}

func (b *Bot) handleLang(ctx context.Context, msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	arg := strings.TrimSpace(msg.CommandArguments())

	// No argument: show current language and usage
	if arg == "" {
		b.send(msg.Chat.ID, l.Get(i18n.MsgLangCurrent)+"\n\n"+l.Get(i18n.MsgLangUsage))
		return
	}

	// Validate language code
	if !i18n.IsValidLang(arg) {
		b.send(msg.Chat.ID, l.Get(i18n.MsgLangUnknown))
		return
	}

	// Save preference to database
	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, l.Get(i18n.MsgFailedVault))
		return
	}

	newLang := i18n.ParseLang(arg)
	if err := vault.SetSetting("language", string(newLang)); err != nil {
		slog.Error("failed to save language setting", "user_id", msg.From.ID, "error", err)
		b.send(msg.Chat.ID, l.Get(i18n.MsgFailedVault))
		return
	}

	// Update cache
	i18n.CacheLang(msg.From.ID, newLang)

	// Confirm in the NEW language
	newLocalizer := i18n.New(string(newLang))
	b.send(msg.Chat.ID, newLocalizer.Get(i18n.MsgLangChanged))
}
