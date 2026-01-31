package ingest

type ContentType string

const (
	ContentTypeLink   ContentType = "link"
	ContentTypeNote   ContentType = "note"
	ContentTypeImage  ContentType = "image"
	ContentTypeSearch ContentType = "search"
)

// ImageFile holds raw image data for a single image.
type ImageFile struct {
	Data []byte
	Ext  string // jpg, png, etc.
}

type RawContent struct {
	Type      ContentType
	URL       string // for links
	Text      string // raw text or note content
	UserID    int64
	ImageData []byte      // raw image bytes (single image, backward compat)
	ImageExt  string      // file extension: jpg, png, etc.
	Images    []ImageFile // multiple images (media group)
	Caption   string      // optional Telegram caption
	Language  string      // user's preferred language code (e.g., "en", "ru")
}

type InputSource interface {
	Name() string
}
