// ABOUTME: HTTP response handling utilities for Gin framework.
// ABOUTME: Provides unified response format and error handling.

package core

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bingo/pkg/contextx"
	"bingo/pkg/errorsx"
)

type ErrResponse struct {
	Reason   string            `json:"reason,omitempty"`
	Message  string            `json:"message,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func Response(c *gin.Context, data any, err error) {
	if err != nil {
		errx := errorsx.FromError(err)
		c.JSON(errx.Code, ErrResponse{
			Reason:   errx.Reason,
			Message:  errx.Message,
			Metadata: errx.Metadata,
		})

		// Set errno to ctx
		contextx.WithMessage(c, errx.Message)

		return
	}

	// 如果没有错误，返回成功响应
	c.JSON(http.StatusOK, data)
}
