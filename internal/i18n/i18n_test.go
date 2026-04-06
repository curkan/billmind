package i18n

import (
	"testing"
)

func TestDetectSystemLang(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected Lang
	}{
		{
			name:     "ru_RU.UTF-8 in LANG",
			envVars:  map[string]string{"LANG": "ru_RU.UTF-8", "LC_ALL": "", "LANGUAGE": ""},
			expected: LangRu,
		},
		{
			name:     "en_US.UTF-8 in LANG",
			envVars:  map[string]string{"LANG": "en_US.UTF-8", "LC_ALL": "", "LANGUAGE": ""},
			expected: LangEn,
		},
		{
			name:     "ru in LC_ALL takes priority",
			envVars:  map[string]string{"LANG": "en_US.UTF-8", "LC_ALL": "ru_RU.UTF-8", "LANGUAGE": ""},
			expected: LangRu,
		},
		{
			name:     "ru in LANGUAGE",
			envVars:  map[string]string{"LANG": "", "LC_ALL": "", "LANGUAGE": "ru"},
			expected: LangRu,
		},
		{
			name:     "empty env vars default to English",
			envVars:  map[string]string{"LANG": "", "LC_ALL": "", "LANGUAGE": ""},
			expected: LangEn,
		},
		{
			name:     "case insensitive detection",
			envVars:  map[string]string{"LANG": "RU_RU.UTF-8", "LC_ALL": "", "LANGUAGE": ""},
			expected: LangRu,
		},
		{
			name:     "fr_FR defaults to English",
			envVars:  map[string]string{"LANG": "fr_FR.UTF-8", "LC_ALL": "", "LANGUAGE": ""},
			expected: LangEn,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for key, val := range tc.envVars {
				t.Setenv(key, val)
			}

			got := DetectSystemLang()
			if got != tc.expected {
				t.Errorf("DetectSystemLang() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestT(t *testing.T) {
	tests := []struct {
		name     string
		lang     Lang
		key      string
		expected string
	}{
		{
			name:     "Russian translation exists",
			lang:     LangRu,
			key:      "help.add",
			expected: "добавить",
		},
		{
			name:     "English translation exists",
			lang:     LangEn,
			key:      "help.add",
			expected: "add",
		},
		{
			name:     "missing key falls back to key itself",
			lang:     LangRu,
			key:      "nonexistent.key",
			expected: "nonexistent.key",
		},
		{
			name:     "missing key in English falls back to key",
			lang:     LangEn,
			key:      "nonexistent.key",
			expected: "nonexistent.key",
		},
		{
			name:     "Russian wizard title",
			lang:     LangRu,
			key:      "wizard.title",
			expected: "Новое напоминание",
		},
		{
			name:     "English wizard title",
			lang:     LangEn,
			key:      "wizard.title",
			expected: "New reminder",
		},
		{
			name:     "status message with format placeholder",
			lang:     LangRu,
			key:      "status.saved",
			expected: "✓ %s сохранён",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SetLang(tc.lang)

			got := T(tc.key)
			if got != tc.expected {
				t.Errorf("T(%q) = %q, want %q", tc.key, got, tc.expected)
			}
		})
	}
}

func TestSetLang(t *testing.T) {
	tests := []struct {
		name string
		lang Lang
	}{
		{name: "set to Russian", lang: LangRu},
		{name: "set to English", lang: LangEn},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SetLang(tc.lang)

			got := CurrentLang()
			if got != tc.lang {
				t.Errorf("CurrentLang() = %q, want %q", got, tc.lang)
			}
		})
	}
}

func TestTFallbackToEnglish(t *testing.T) {
	// Ensure that if a key exists only in English and not in Russian,
	// it falls back to the English translation.
	// We temporarily add a key only to English to test this.
	original := translations[LangEn]["help.add"]

	// Add a test-only key to English
	translations[LangEn]["test.only.en"] = "english only"

	SetLang(LangRu)
	got := T("test.only.en")
	if got != "english only" {
		t.Errorf("T(%q) with LangRu = %q, want %q (fallback to English)", "test.only.en", got, "english only")
	}

	// Cleanup
	delete(translations[LangEn], "test.only.en")

	// Verify original is unchanged
	if translations[LangEn]["help.add"] != original {
		t.Error("original translations were modified")
	}
}

func TestAllKeysExistInBothLanguages(t *testing.T) {
	t.Parallel()

	ruKeys := translations[LangRu]
	enKeys := translations[LangEn]

	for key := range ruKeys {
		if _, ok := enKeys[key]; !ok {
			t.Errorf("key %q exists in Russian but not in English", key)
		}
	}

	for key := range enKeys {
		if _, ok := ruKeys[key]; !ok {
			t.Errorf("key %q exists in English but not in Russian", key)
		}
	}
}
