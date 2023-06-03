package core

import (
	"net/http"

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
	if err != nil {
		hcode, code, message := errno.Decode(err)
		c.JSON(hcode, ErrResponse{
			Code:    code,
			Message: message,
		})

		return
	}

	c.JSON(http.StatusOK, data)
}
