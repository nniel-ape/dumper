package bot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	default:
		b.send(msg.Chat.ID, "Unknown command. Use /help to see available commands.")
	}
}

func (b *Bot) handleStart(msg *tgbotapi.Message) {
	text := `👋 <b>Welcome to Dumper!</b>

I help you capture and organize knowledge from the web.

<b>How to use:</b>
• Send me any link - I'll extract, summarize, and tag it
• Send me text notes - I'll categorize them too
• Use /search to find saved items
• Use /recent to see your latest items
• Use /tags to see all your tags

All your data is stored privately and can be exported to Obsidian.`

	if b.webAppURL != "" {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("📱 Open App", b.webAppURL),
			),
		)
		b.sendWithKeyboard(msg.Chat.ID, text, keyboard)
	} else {
		b.send(msg.Chat.ID, text)
	}
}

func (b *Bot) handleHelp(msg *tgbotapi.Message) {
	text := `<b>Commands:</b>
/search [query] - Search your saved items
/recent - Show recent items
/tags - List all your tags
/stats - Show vault statistics
/export - Export to Obsidian format
/app - Open Mini App (if configured)

<b>Saving content:</b>
Just send me any URL or text message!`
	b.send(msg.Chat.ID, text)
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

	var raw ingest.RawContent
	raw.UserID = msg.From.ID

	if ingest.IsURL(text) {
		b.send(msg.Chat.ID, "⏳ Processing link...")
		raw.Type = ingest.ContentTypeLink
		raw.URL = text
	} else if ingest.IsShortTopicMessage(text) {
		b.send(msg.Chat.ID, fmt.Sprintf("🔍 Searching: <b>%s</b>...", text))
		raw.Type = ingest.ContentTypeSearch
		raw.Text = text
	} else {
		b.send(msg.Chat.ID, "⏳ Processing note...")
		raw.Type = ingest.ContentTypeNote
		raw.Text = text
	}

	item, err := b.pipeline.Process(ctx, raw)
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to process: %v", err))
		return
	}

	// Format response
	var tagsStr string
	if len(item.Tags) > 0 {
		tagsStr = "#" + strings.Join(item.Tags, " #")
	}

	response := fmt.Sprintf(`✅ <b>Saved!</b>

<b>%s</b>

%s

%s`, item.Title, item.Summary, tagsStr)

	if b.webAppURL != "" {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("View in App", b.webAppURL+"?item="+item.ID),
			),
		)
		b.sendWithKeyboard(msg.Chat.ID, response, keyboard)
	} else {
		b.send(msg.Chat.ID, response)
	}
}

func (b *Bot) handlePhoto(ctx context.Context, msg *tgbotapi.Message) {
	// Get largest photo (last in slice)
	photos := msg.Photo
	photo := photos[len(photos)-1]

	b.send(msg.Chat.ID, "📷 Saving image...")

	// Get file info from Telegram
	fileConfig := tgbotapi.FileConfig{FileID: photo.FileID}
	file, err := b.api.GetFile(fileConfig)
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to get file info: %v", err))
		return
	}

	// Download file
	fileURL := file.Link(b.api.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to download image: %v", err))
		return
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to read image: %v", err))
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
	}

	item, err := b.pipeline.Process(ctx, raw)
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to save image: %v", err))
		return
	}

	// Format response
	var tagsStr string
	if len(item.Tags) > 0 {
		tagsStr = "#" + strings.Join(item.Tags, " #")
	}

	response := fmt.Sprintf(`✅ <b>Image saved!</b>

<b>%s</b>

%s`, item.Title, tagsStr)

	if b.webAppURL != "" {
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonURL("View in App", b.webAppURL+"?item="+item.ID),
			),
		)
		b.sendWithKeyboard(msg.Chat.ID, response, keyboard)
	} else {
		b.send(msg.Chat.ID, response)
	}
}

func (b *Bot) handleSearch(ctx context.Context, msg *tgbotapi.Message) {
	query := msg.CommandArguments()
	if query == "" {
		b.send(msg.Chat.ID, "Usage: /search [query]\nExample: /search golang concurrency")
		return
	}

	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, "❌ Failed to access your vault")
		return
	}

	results, err := vault.Search(query, 5)
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Search failed: %v", err))
		return
	}

	if len(results) == 0 {
		b.send(msg.Chat.ID, "No results found.")
		return
	}

	var text strings.Builder
	text.WriteString(fmt.Sprintf("🔍 <b>Results for \"%s\":</b>\n\n", query))

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
	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, "❌ Failed to access your vault")
		return
	}

	items, err := vault.ListItems(5, 0)
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to list items: %v", err))
		return
	}

	if len(items) == 0 {
		b.send(msg.Chat.ID, "No items saved yet. Send me a link or note to get started!")
		return
	}

	var text strings.Builder
	text.WriteString("📚 <b>Recent items:</b>\n\n")

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
	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, "❌ Failed to access your vault")
		return
	}

	tags, err := vault.GetAllTags()
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to get tags: %v", err))
		return
	}

	if len(tags) == 0 {
		b.send(msg.Chat.ID, "No tags yet.")
		return
	}

	b.send(msg.Chat.ID, fmt.Sprintf("🏷 <b>Your tags:</b>\n\n#%s", strings.Join(tags, " #")))
}

func (b *Bot) handleStats(ctx context.Context, msg *tgbotapi.Message) {
	vault, err := b.stores.GetVault(msg.From.ID)
	if err != nil {
		b.send(msg.Chat.ID, "❌ Failed to access your vault")
		return
	}

	count, err := vault.ItemCount()
	if err != nil {
		b.send(msg.Chat.ID, fmt.Sprintf("❌ Failed to get stats: %v", err))
		return
	}

	tags, _ := vault.GetAllTags()

	b.send(msg.Chat.ID, fmt.Sprintf("📊 <b>Your vault:</b>\n\n• Items: %d\n• Tags: %d", count, len(tags)))
}

func (b *Bot) handleExport(ctx context.Context, msg *tgbotapi.Message) {
	b.send(msg.Chat.ID, "Export feature coming soon! Use the API endpoint /api/export for now.")
}

func (b *Bot) handleApp(msg *tgbotapi.Message) {
	if b.webAppURL == "" {
		b.send(msg.Chat.ID, "Mini App is not configured. Set WEBAPP_URL environment variable.")
		return
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📱 Open Mini App", b.webAppURL),
		),
	)
	b.sendWithKeyboard(msg.Chat.ID, "Open the Mini App to browse, search, and visualize your knowledge:", keyboard)
}
