// ABOUTME: Language middleware for extracting Accept-Language from requests.
// ABOUTME: Sets the language preference in context for i18n support.

package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/pkg/contextx"
)

// DefaultLang is the default language when Accept-Language is not specified.
const DefaultLang = "en"

// Lang extracts Accept-Language header and stores it in context.
func Lang() gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := parseAcceptLanguage(c.GetHeader("Accept-Language"))
		ctx := contextx.WithLang(c.Request.Context(), lang)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// parseAcceptLanguage extracts the primary language from Accept-Language header.
// Examples:
//   - "zh-CN,zh;q=0.9,en;q=0.8" -> "zh"
//   - "en-US" -> "en"
//   - "" -> "en" (default)
func parseAcceptLanguage(header string) string {
	if header == "" {
		return DefaultLang
	}

	// Take the first language preference
	parts := strings.Split(header, ",")
	if len(parts) == 0 {
		return DefaultLang
	}

	// Remove quality factor (q=...) and region code
	lang := strings.TrimSpace(parts[0])
	if idx := strings.Index(lang, ";"); idx > 0 {
		lang = lang[:idx]
	}

	// Extract base language (e.g., "zh-CN" -> "zh")
	if idx := strings.Index(lang, "-"); idx > 0 {
		lang = lang[:idx]
	}

	if lang == "" {
		return DefaultLang
	}

	return strings.ToLower(lang)
}
