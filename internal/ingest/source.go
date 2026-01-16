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
}

type InputSource interface {
	Name() string
}
