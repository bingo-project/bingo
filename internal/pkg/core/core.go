package core

import (
	"net/http"

	"github.com/bingo-project/component-base/log"
	"github.com/gin-gonic/gin"

	"bingo/internal/pkg/errno"
)

// ErrResponse defines the return messages when an error occurred.
type ErrResponse struct {
	// Code defines the business error code.
	Code string `json:"code"`

	// Message contains the detail of this message.
	// This message is suitable to be exposed to external
	Message string `json:"message"`
}

// WriteResponse write an error or the response data into http response body.
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
