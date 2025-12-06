// ABOUTME: HTTP response handling utilities for Gin framework.
// ABOUTME: Provides unified response format and error handling.
package core

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/errno"
	"bingo/internal/pkg/log"
	"bingo/pkg/errorsx"
)

// ErrResponse defines the return messages when an error occurred.
// Deprecated: Use ErrorResponse instead, which uses errorsx for error handling.
type ErrResponse struct {
	// Code defines the business error code.
	Code string `json:"code"`

	// Message contains the detail of this message.
	// This message is suitable to be exposed to external
	Message string `json:"message"`
}

// WriteResponse write an error or the response data into http response body.
// Deprecated: Use Response instead, which uses errorsx for unified error handling.
func WriteResponse(c *gin.Context, err error, data interface{}) {
	hcode, code, message := errno.Decode(err)

	// Set errno to ctx
	c.Set(log.KeyCode, code)
	c.Set(log.KeyMessage, message)

	if err != nil {
		c.JSON(hcode, ErrResponse{
			Code:    code,
			Message: message,
		})

		return
	}

	c.JSON(http.StatusOK, data)
}

type ErrorResponse struct {
	Reason   string            `json:"reason,omitempty"`
	Message  string            `json:"message,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func Response(c *gin.Context, data any, err error) {
	if err != nil {
		errx := errorsx.FromError(err)
		c.JSON(errx.Code, ErrorResponse{
			Reason:   errx.Reason,
			Message:  errx.Message,
			Metadata: errx.Metadata,
		})

		return
	}

	// 如果没有错误，返回成功响应
	c.JSON(http.StatusOK, data)
}
