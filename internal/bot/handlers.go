package bot

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strings"

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

	if ingest.IsURL(text) {
		b.send(msg.Chat.ID, l.Get(i18n.MsgProcessingLink))
		raw.Type = ingest.ContentTypeLink
		raw.URL = text
	} else if ingest.IsShortTopicMessage(text) {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgSearching, text))
		raw.Type = ingest.ContentTypeSearch
		raw.Text = text
	} else {
		b.send(msg.Chat.ID, l.Get(i18n.MsgProcessingNote))
		raw.Type = ingest.ContentTypeNote
		raw.Text = text
	}

	item, err := b.pipeline.Process(ctx, raw)
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedProcess, err))
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
		b.sendWithKeyboard(msg.Chat.ID, response, keyboard)
	} else {
		b.send(msg.Chat.ID, response)
	}
}

func (b *Bot) handlePhoto(ctx context.Context, msg *tgbotapi.Message) {
	l := b.getUserLang(msg.From.ID, msg.From.LanguageCode)

	// Get largest photo (last in slice)
	photos := msg.Photo
	photo := photos[len(photos)-1]

	b.send(msg.Chat.ID, l.Get(i18n.MsgSavingImage))

	// Get file info from Telegram
	fileConfig := tgbotapi.FileConfig{FileID: photo.FileID}
	file, err := b.api.GetFile(fileConfig)
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedFileInfo, err))
		return
	}

	// Download file
	fileURL := file.Link(b.api.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedDownload, err))
		return
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedReadImage, err))
		return
	}

	// Determine file extension
	ext := "jpg" // default
	if filePath := file.FilePath; filePath != "" {
		if e := path.Ext(filePath); e != "" {
			ext = strings.TrimPrefix(e, ".")
		}
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
		b.send(msg.Chat.ID, l.Getf(i18n.MsgFailedSaveImage, err))
		return
	}

	// Format response
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
		b.sendWithKeyboard(msg.Chat.ID, response, keyboard)
	} else {
		b.send(msg.Chat.ID, response)
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

	tags, _ := vault.GetAllTags()

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
