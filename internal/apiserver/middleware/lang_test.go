// ABOUTME: Tests for language middleware.
// ABOUTME: Verifies Accept-Language header parsing and context storage.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/bingo-project/bingo/pkg/contextx"
)

func TestLangMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		acceptLanguage string
		expectedLang   string
	}{
		{
			name:           "Chinese with region",
			acceptLanguage: "zh-CN,zh;q=0.9,en;q=0.8",
			expectedLang:   "zh",
		},
		{
			name:           "English with region",
			acceptLanguage: "en-US",
			expectedLang:   "en",
		},
		{
			name:           "Simple Chinese",
			acceptLanguage: "zh",
			expectedLang:   "zh",
		},
		{
			name:           "Empty header defaults to English",
			acceptLanguage: "",
			expectedLang:   "en",
		},
		{
			name:           "Japanese",
			acceptLanguage: "ja-JP,ja;q=0.9",
			expectedLang:   "ja",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedLang string

			router := gin.New()
			router.Use(Lang())
			router.GET("/test", func(c *gin.Context) {
				capturedLang = contextx.Lang(c.Request.Context())
				c.Status(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.acceptLanguage != "" {
				req.Header.Set("Accept-Language", tt.acceptLanguage)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, tt.expectedLang, capturedLang)
		})
	}
}

func TestParseAcceptLanguage(t *testing.T) {
	tests := []struct {
		header   string
		expected string
	}{
		{"zh-CN,zh;q=0.9,en;q=0.8", "zh"},
		{"en-US", "en"},
		{"zh", "zh"},
		{"", "en"},
		{"fr-FR;q=0.9", "fr"},
		{"de-DE,de;q=0.8", "de"},
	}

	for _, tt := range tests {
		t.Run(tt.header, func(t *testing.T) {
			result := parseAcceptLanguage(tt.header)
			assert.Equal(t, tt.expected, result)
		})
	}
}
