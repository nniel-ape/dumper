package i18n

// MsgKey is a typed key for translation messages.
type MsgKey string

// Bot messages
const (
	// Welcome and help
	MsgWelcome        MsgKey = "welcome"
	MsgHelp           MsgKey = "help"
	MsgUnknownCommand MsgKey = "unknown_command"

	// Processing status
	MsgProcessingLink  MsgKey = "processing_link"
	MsgProcessingNote  MsgKey = "processing_note"
	MsgSavingImage     MsgKey = "saving_image"
	MsgSearching       MsgKey = "searching"
	MsgSearchUsage     MsgKey = "search_usage"
	MsgRecentItems     MsgKey = "recent_items"
	MsgYourTags        MsgKey = "your_tags"
	MsgYourVault       MsgKey = "your_vault"
	MsgOpenApp         MsgKey = "open_app"
	MsgViewInApp       MsgKey = "view_in_app"
	MsgAppNotConfigured MsgKey = "app_not_configured"
	MsgOpenMiniApp     MsgKey = "open_mini_app"
	MsgExportComingSoon MsgKey = "export_coming_soon"

	// Success messages
	MsgSaved      MsgKey = "saved"
	MsgImageSaved MsgKey = "image_saved"

	// Empty states
	MsgNoResults  MsgKey = "no_results"
	MsgNoItems    MsgKey = "no_items"
	MsgNoTags     MsgKey = "no_tags"
	MsgSearchFor  MsgKey = "search_for"

	// Errors
	MsgFailedProcess   MsgKey = "failed_process"
	MsgFailedVault     MsgKey = "failed_vault"
	MsgFailedSearch    MsgKey = "failed_search"
	MsgFailedListItems MsgKey = "failed_list_items"
	MsgFailedGetTags   MsgKey = "failed_get_tags"
	MsgFailedGetStats  MsgKey = "failed_get_stats"
	MsgFailedFileInfo  MsgKey = "failed_file_info"
	MsgFailedDownload  MsgKey = "failed_download"
	MsgFailedReadImage MsgKey = "failed_read_image"
	MsgFailedSaveImage MsgKey = "failed_save_image"

	// Language
	MsgLangCurrent MsgKey = "lang_current"
	MsgLangUsage   MsgKey = "lang_usage"
	MsgLangChanged MsgKey = "lang_changed"
	MsgLangUnknown MsgKey = "lang_unknown"
)
