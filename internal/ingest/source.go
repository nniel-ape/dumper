package ingest

type ContentType string

const (
	ContentTypeLink ContentType = "link"
	ContentTypeNote ContentType = "note"
)

type RawContent struct {
	Type   ContentType
	URL    string // for links
	Text   string // raw text or note content
	UserID int64
}

type InputSource interface {
	Name() string
}
