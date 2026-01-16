package ingest

type ContentType string

const (
	ContentTypeLink   ContentType = "link"
	ContentTypeNote   ContentType = "note"
	ContentTypeImage  ContentType = "image"
	ContentTypeSearch ContentType = "search"
)

type RawContent struct {
	Type      ContentType
	URL       string // for links
	Text      string // raw text or note content
	UserID    int64
	ImageData []byte // raw image bytes (for images)
	ImageExt  string // file extension: jpg, png, etc.
	Caption   string // optional Telegram caption
	Language  string // user's preferred language code (e.g., "en", "ru")
}

type InputSource interface {
	Name() string
}
