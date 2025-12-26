// ABOUTME: Internationalization (i18n) support for the application.
// ABOUTME: Provides translation functions with embedded locale files.

package i18n

import (
	"embed"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed locales/*.yaml
var localeFS embed.FS

var (
	bundle *i18n.Bundle
	once   sync.Once
)

// DefaultLang is the default language when no language is specified.
const DefaultLang = "en"

// Init initializes the i18n bundle with embedded locale files.
func Init() {
	once.Do(func() {
		bundle = i18n.NewBundle(language.English)
		bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

		// Load embedded locale files
		_, _ = bundle.LoadMessageFileFS(localeFS, "locales/en.yaml")
		_, _ = bundle.LoadMessageFileFS(localeFS, "locales/zh.yaml")
	})
}

// T translates a message ID to the specified language.
// If data is nil, no template substitution is performed.
func T(lang, messageID string, data map[string]interface{}) string {
	if bundle == nil {
		Init()
	}

	if lang == "" {
		lang = DefaultLang
	}

	localizer := i18n.NewLocalizer(bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		// Fallback to message ID if translation not found
		return messageID
	}

	return msg
}

// TWithDefault translates a message ID, returning defaultMsg if not found.
func TWithDefault(lang, messageID, defaultMsg string, data map[string]interface{}) string {
	if bundle == nil {
		Init()
	}

	if lang == "" {
		lang = DefaultLang
	}

	localizer := i18n.NewLocalizer(bundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return defaultMsg
	}

	return msg
}
