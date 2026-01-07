// ABOUTME: AI module error codes.
// ABOUTME: Defines errors for chat, session, quota operations.

package errno

import (
	"net/http"

	"github.com/bingo-project/bingo/pkg/errorsx"
)

var (
	// ErrAIModelNotFound 模型不存在
	ErrAIModelNotFound = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AIModelNotFound",
		Message: "AI model not found.",
	}

	// ErrAIProviderNotFound Provider 不存在
	ErrAIProviderNotFound = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AIProviderNotFound",
		Message: "AI provider not found.",
	}

	// ErrAIProviderNotConfigured Provider 未配置
	ErrAIProviderNotConfigured = &errorsx.ErrorX{
		Code:    http.StatusServiceUnavailable,
		Reason:  "InternalError.AIProviderNotConfigured",
		Message: "AI provider is not configured.",
	}

	// ErrAIQuotaExceeded 配额超限
	ErrAIQuotaExceeded = &errorsx.ErrorX{
		Code:    http.StatusTooManyRequests,
		Reason:  "ResourceExhausted.AIQuotaExceeded",
		Message: "AI quota exceeded.",
	}

	// ErrAISessionNotFound 会话不存在
	ErrAISessionNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.AISessionNotFound",
		Message: "AI session not found.",
	}

	// ErrAIStreamError 流式响应错误
	ErrAIStreamError = &errorsx.ErrorX{
		Code:    http.StatusInternalServerError,
		Reason:  "InternalError.AIStreamError",
		Message: "AI stream error.",
	}

	// ErrAIProviderError Provider 返回错误
	ErrAIProviderError = &errorsx.ErrorX{
		Code:    http.StatusBadGateway,
		Reason:  "ExternalError.AIProviderError",
		Message: "AI provider returned an error.",
	}

	// ErrAIContextTooLong 上下文过长
	ErrAIContextTooLong = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AIContextTooLong",
		Message: "AI context is too long.",
	}

	// ErrAIEmptyMessages 消息为空
	ErrAIEmptyMessages = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AIEmptyMessages",
		Message: "Messages cannot be empty.",
	}

	// ErrAIMessageTooLong 消息过长
	ErrAIMessageTooLong = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AIMessageTooLong",
		Message: "Message content is too long.",
	}

	// ErrAIRoleNotFound 角色不存在
	ErrAIRoleNotFound = &errorsx.ErrorX{
		Code:    http.StatusNotFound,
		Reason:  "NotFound.AIRoleNotFound",
		Message: "AI role not found.",
	}

	// ErrAIRoleDisabled 角色已禁用
	ErrAIRoleDisabled = &errorsx.ErrorX{
		Code:    http.StatusBadRequest,
		Reason:  "InvalidArgument.AIRoleDisabled",
		Message: "AI role is disabled.",
	}

	// ErrAIAllModelsFailed 所有模型（包括降级）都失败
	ErrAIAllModelsFailed = &errorsx.ErrorX{
		Code:    http.StatusServiceUnavailable,
		Reason:  "ServiceUnavailable.AllModelsFailed",
		Message: "AI service is temporarily unavailable, please try again later.",
	}
)
