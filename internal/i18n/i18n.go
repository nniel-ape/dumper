package i18n

import (
	"fmt"
	"strings"
	"sync"
)

// Lang represents a supported language.
type Lang string

const (
	LangEN Lang = "en"
	LangRU Lang = "ru"
)

// translations maps language to message map.
var translations = map[Lang]map[MsgKey]string{
	LangEN: en,
	LangRU: ru,
}

// Localizer provides localized messages for a specific language.
type Localizer struct {
	lang Lang
}

// New creates a Localizer for the given language code.
func New(code string) *Localizer {
	return &Localizer{lang: ParseLang(code)}
}

// Lang returns the current language.
func (l *Localizer) Lang() Lang {
	return l.lang
}

// Code returns the language code as string.
func (l *Localizer) Code() string {
	return string(l.lang)
}

// Get returns the localized message for the given key.
func (l *Localizer) Get(key MsgKey) string {
	if msgs, ok := translations[l.lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	// Fallback to English
	if msgs, ok := translations[LangEN]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	return string(key)
}

// Getf returns the localized message with fmt.Sprintf formatting.
func (l *Localizer) Getf(key MsgKey, args ...any) string {
	return fmt.Sprintf(l.Get(key), args...)
}

// ParseLang converts a language code to a Lang.
// Supports: "ru", "ru-RU", "uk" (Ukrainian), "be" (Belarusian) -> LangRU
// Everything else -> LangEN
func ParseLang(code string) Lang {
	code = strings.ToLower(strings.TrimSpace(code))

	// Extract primary language tag (before hyphen)
	if idx := strings.Index(code, "-"); idx > 0 {
		code = code[:idx]
	}
	if idx := strings.Index(code, "_"); idx > 0 {
		code = code[:idx]
	}

	switch code {
	case "ru", "uk", "be":
		// Russian, Ukrainian, Belarusian -> Russian UI
		return LangRU
	default:
		return LangEN
	}
}

// --- Memory cache for per-user language preferences ---

var (
	langCache   = make(map[int64]Lang)
	langCacheMu sync.RWMutex
)

// CacheLang stores a user's language preference in memory.
func CacheLang(userID int64, lang Lang) {
	langCacheMu.Lock()
	langCache[userID] = lang
	langCacheMu.Unlock()
}

// GetCachedLang retrieves a user's cached language preference.
func GetCachedLang(userID int64) (Lang, bool) {
	langCacheMu.RLock()
	lang, ok := langCache[userID]
	langCacheMu.RUnlock()
	return lang, ok
}

// ClearCachedLang removes a user's language from cache.
func ClearCachedLang(userID int64) {
	langCacheMu.Lock()
	delete(langCache, userID)
	langCacheMu.Unlock()
}

// IsValidLang checks if a language code is supported.
func IsValidLang(code string) bool {
	code = strings.ToLower(strings.TrimSpace(code))
	return code == "en" || code == "ru"
}
