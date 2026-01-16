package i18n

var en = map[MsgKey]string{
	// Welcome and help
	MsgWelcome: `👋 <b>Welcome to Dumper!</b>

I help you capture and organize knowledge from the web.

<b>How to use:</b>
• Send me any link - I'll extract, summarize, and tag it
• Send me text notes - I'll categorize them too
• Use /search to find saved items
• Use /recent to see your latest items
• Use /tags to see all your tags
• Use /lang to change language

All your data is stored privately and can be exported to Obsidian.`,

	MsgHelp: `<b>Commands:</b>
/search [query] - Search your saved items
/recent - Show recent items
/tags - List all your tags
/stats - Show vault statistics
/export - Export to Obsidian format
/app - Open Mini App (if configured)
/lang - Change language (en/ru)

<b>Saving content:</b>
Just send me any URL or text message!`,

	MsgUnknownCommand: "Unknown command. Use /help to see available commands.",

	// Processing status
	MsgProcessingLink:   "⏳ Processing link...",
	MsgProcessingNote:   "⏳ Processing note...",
	MsgSavingImage:      "📷 Saving image...",
	MsgSearching:        "🔍 Searching: <b>%s</b>...",
	MsgSearchUsage:      "Usage: /search [query]\nExample: /search golang concurrency",
	MsgRecentItems:      "📚 <b>Recent items:</b>\n\n",
	MsgYourTags:         "🏷 <b>Your tags:</b>\n\n#%s",
	MsgYourVault:        "📊 <b>Your vault:</b>\n\n• Items: %d\n• Tags: %d",
	MsgOpenApp:          "📱 Open App",
	MsgViewInApp:        "View in App",
	MsgAppNotConfigured: "Mini App is not configured. Set WEBAPP_URL environment variable.",
	MsgOpenMiniApp:      "Open the Mini App to browse, search, and visualize your knowledge:",
	MsgExportComingSoon: "Export feature coming soon! Use the API endpoint /api/export for now.",

	// Success messages
	MsgSaved:      "✅ <b>Saved!</b>",
	MsgImageSaved: "✅ <b>Image saved!</b>",

	// Empty states
	MsgNoResults: "No results found.",
	MsgNoItems:   "No items saved yet. Send me a link or note to get started!",
	MsgNoTags:    "No tags yet.",
	MsgSearchFor: `🔍 <b>Results for "%s":</b>`,

	// Errors
	MsgFailedProcess:   "❌ Failed to process: %v",
	MsgFailedVault:     "❌ Failed to access your vault",
	MsgFailedSearch:    "❌ Search failed: %v",
	MsgFailedListItems: "❌ Failed to list items: %v",
	MsgFailedGetTags:   "❌ Failed to get tags: %v",
	MsgFailedGetStats:  "❌ Failed to get stats: %v",
	MsgFailedFileInfo:  "❌ Failed to get file info: %v",
	MsgFailedDownload:  "❌ Failed to download image: %v",
	MsgFailedReadImage: "❌ Failed to read image: %v",
	MsgFailedSaveImage: "❌ Failed to save image: %v",

	// Language
	MsgLangCurrent: "🌐 Current language: <b>English</b>\n\nUse /lang ru to switch to Russian.",
	MsgLangUsage:   "Usage: /lang [en|ru]\n\nAvailable languages:\n• en - English\n• ru - Русский",
	MsgLangChanged: "✅ Language changed to <b>English</b>",
	MsgLangUnknown: "Unknown language. Available: en, ru",
}
