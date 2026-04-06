package i18n

import (
	"os"
	"strings"
	"sync"
)

// Lang represents a supported language.
type Lang string

const (
	LangRu Lang = "ru"
	LangEn Lang = "en"
)

var (
	current Lang = LangRu
	mu      sync.RWMutex
)

// SetLang sets the current application language.
func SetLang(lang Lang) {
	mu.Lock()
	defer mu.Unlock()
	current = lang
}

// CurrentLang returns the current application language.
func CurrentLang() Lang {
	mu.RLock()
	defer mu.RUnlock()
	return current
}

// T returns the translated string for the given key using the current language.
// Falls back to English translation if not found in the current language.
// Falls back to the key itself if not found in any language.
func T(key string) string {
	mu.RLock()
	lang := current
	mu.RUnlock()

	if dict, ok := translations[lang]; ok {
		if val, ok := dict[key]; ok {
			return val
		}
	}

	if lang != LangEn {
		if dict, ok := translations[LangEn]; ok {
			if val, ok := dict[key]; ok {
				return val
			}
		}
	}

	return key
}

// DetectSystemLang detects the user's language from environment variables
// LANG, LC_ALL, and LANGUAGE. If any contains "ru", returns LangRu.
// Otherwise returns LangEn.
func DetectSystemLang() Lang {
	for _, env := range []string{"LC_ALL", "LANG", "LANGUAGE"} {
		val := os.Getenv(env)
		if val == "" {
			continue
		}
		lower := strings.ToLower(val)
		if strings.Contains(lower, "ru") {
			return LangRu
		}
	}
	return LangEn
}
