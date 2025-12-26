// ABOUTME: SMS service interface for sending verification codes.
// ABOUTME: Provides a placeholder implementation until actual SMS provider is configured.

package sms

import "github.com/bingo-project/bingo/internal/pkg/errno"

// SMS 短信发送接口
type SMS interface {
	Send(phone string, content string) error
}

// nopSMS 空实现（未配置时使用）
type nopSMS struct{}

// NewNopSMS 创建空实现
func NewNopSMS() SMS {
	return &nopSMS{}
}

func (n *nopSMS) Send(phone, content string) error {
	return errno.ErrSMSNotConfigured
}

// IsConfigured 检查 SMS 是否已配置
// TODO: 实际接入时根据配置判断
func IsConfigured() bool {
	return false
}
