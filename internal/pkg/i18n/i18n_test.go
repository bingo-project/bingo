// ABOUTME: Tests for i18n translation functionality.
// ABOUTME: Verifies translation loading and message rendering.

package i18n

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestT_English(t *testing.T) {
	Init()

	tests := []struct {
		name      string
		lang      string
		messageID string
		data      map[string]interface{}
		expected  string
	}{
		{
			name:      "register subject in English",
			lang:      "en",
			messageID: "code_register_subject",
			data:      nil,
			expected:  "Registration Verification Code",
		},
		{
			name:      "register body in English with data",
			lang:      "en",
			messageID: "code_register_body",
			data:      map[string]interface{}{"Code": "123456", "TTL": 5},
			expected:  "You are registering an account. Your verification code is: 123456. Valid for 5 minutes.",
		},
		{
			name:      "reset password subject in English",
			lang:      "en",
			messageID: "code_reset_password_subject",
			data:      nil,
			expected:  "Password Reset Verification Code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := T(tt.lang, tt.messageID, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestT_Chinese(t *testing.T) {
	Init()

	tests := []struct {
		name      string
		lang      string
		messageID string
		data      map[string]interface{}
		expected  string
	}{
		{
			name:      "register subject in Chinese",
			lang:      "zh",
			messageID: "code_register_subject",
			data:      nil,
			expected:  "注册验证码",
		},
		{
			name:      "register body in Chinese with data",
			lang:      "zh",
			messageID: "code_register_body",
			data:      map[string]interface{}{"Code": "123456", "TTL": 5},
			expected:  "您正在注册账号，验证码：123456，5分钟内有效。",
		},
		{
			name:      "reset password subject in Chinese",
			lang:      "zh",
			messageID: "code_reset_password_subject",
			data:      nil,
			expected:  "密码重置验证码",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := T(tt.lang, tt.messageID, tt.data)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestT_DefaultLanguage(t *testing.T) {
	Init()

	// Empty lang should fallback to English
	result := T("", "code_register_subject", nil)
	assert.Equal(t, "Registration Verification Code", result)
}

func TestT_FallbackToMessageID(t *testing.T) {
	Init()

	// Unknown message ID should return the message ID itself
	result := T("en", "unknown_message_id", nil)
	assert.Equal(t, "unknown_message_id", result)
}

func TestTWithDefault(t *testing.T) {
	Init()

	// Known message
	result := TWithDefault("en", "code_register_subject", "default", nil)
	assert.Equal(t, "Registration Verification Code", result)

	// Unknown message should return default
	result = TWithDefault("en", "unknown_message", "my default", nil)
	assert.Equal(t, "my default", result)
}
