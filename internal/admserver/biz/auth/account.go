// ABOUTME: Account type detection for multi-auth system.
// ABOUTME: Provides email/phone format validation and automatic type detection.

package auth

import (
	"regexp"
	"strings"

	"github.com/bingo-project/bingo/internal/pkg/errno"
)

// AccountType 账号类型
type AccountType string

const (
	AccountTypeEmail AccountType = "email"
	AccountTypePhone AccountType = "phone"
)

var (
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	phoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`) // 中国手机号格式
)

// DetectAccountType 自动检测账号类型
func DetectAccountType(account string) (AccountType, error) {
	account = strings.TrimSpace(account)
	if account == "" {
		return "", errno.ErrInvalidAccountFormat
	}

	// 包含 @ 且符合邮箱格式 → email
	if strings.Contains(account, "@") && emailRegex.MatchString(account) {
		return AccountTypeEmail, nil
	}

	// 符合手机号格式 → phone
	if phoneRegex.MatchString(account) {
		return AccountTypePhone, nil
	}

	return "", errno.ErrInvalidAccountFormat
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// IsValidPhone 验证手机号格式
func IsValidPhone(phone string) bool {
	return phoneRegex.MatchString(phone)
}
