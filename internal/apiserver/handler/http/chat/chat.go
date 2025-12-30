// ABOUTME: Chat HTTP handlers for AI chat completions.
// ABOUTME: Provides endpoints for chat, models listing.

package chat

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
)

type ChatHandler struct {
	b biz.IBiz
}

func NewChatHandler(ds store.IStore, registry *ai.Registry) *ChatHandler {
	return &ChatHandler{
		b: biz.NewBiz(ds).WithRegistry(registry),
	}
}

// ChatCompletions
// @Summary    Create chat completion
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.ChatCompletionRequest  true  "Chat request"
// @Success    200      {object}  v1.ChatCompletionResponse
// @Failure    400      {object}  core.ErrResponse
// @Failure    429      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/chat/completions [POST].
func (h *ChatHandler) ChatCompletions(c *gin.Context) {
	var req v1.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	uid := contextx.UserID(c)

	// Convert DTO to ai.ChatRequest
	aiReq := &ai.ChatRequest{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		Stream:      req.Stream,
		SessionID:   req.SessionID,
		UID:         uid,
	}
	for _, msg := range req.Messages {
		aiReq.Messages = append(aiReq.Messages, ai.Message{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	if req.Stream {
		h.handleStream(c, uid, aiReq)

		return
	}

	resp, err := h.b.Chat().Chat(c, uid, aiReq)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, convertToDTO(resp), nil)
}

// handleStream handles streaming response
func (h *ChatHandler) handleStream(c *gin.Context, uid string, req *ai.ChatRequest) {
	stream, err := h.b.Chat().ChatStream(c, uid, req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	c.Stream(func(w io.Writer) bool {
		chunk, err := stream.Recv()
		if err != nil {
			if !errors.Is(err, ai.ErrStreamClosed) {
				// Send error event before DONE for non-normal termination
				errData := map[string]interface{}{
					"error": map[string]string{
						"message": "stream error occurred",
						"type":    "stream_error",
					},
				}
				data, _ := json.Marshal(errData)
				fmt.Fprintf(w, "data: %s\n\n", data)
			}
			fmt.Fprintf(w, "data: [DONE]\n\n")

			return false
		}

		data, _ := json.Marshal(convertChunkToDTO(chunk))
		fmt.Fprintf(w, "data: %s\n\n", data)

		return true
	})
}

// ListModels
// @Summary    List available models
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Success    200  {object}  v1.ListModelsResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/models [GET].
func (h *ChatHandler) ListModels(c *gin.Context) {
	models, err := h.b.Chat().ListModels(c)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	data := make([]v1.ModelInfo, len(models))
	for i, m := range models {
		data[i] = v1.ModelInfo{
			ID:          m.ID,
			Object:      "model",
			Created:     time.Now().Unix(),
			OwnedBy:     m.Provider,
			MaxTokens:   m.MaxTokens,
			InputPrice:  m.InputPrice,
			OutputPrice: m.OutputPrice,
		}
	}

	core.Response(c, v1.ListModelsResponse{
		Object: "list",
		Data:   data,
	}, nil)
}

// convertToDTO converts ai.ChatResponse to v1.ChatCompletionResponse
func convertToDTO(resp *ai.ChatResponse) *v1.ChatCompletionResponse {
	choices := make([]v1.ChatChoice, len(resp.Choices))
	for i, ch := range resp.Choices {
		choices[i] = v1.ChatChoice{
			Index: ch.Index,
			Message: v1.ChatMessage{
				Role:    ch.Message.Role,
				Content: ch.Message.Content,
			},
			FinishReason: ch.FinishReason,
		}
	}

	return &v1.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: v1.ChatUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}

// convertChunkToDTO converts ai.StreamChunk to v1.ChatCompletionResponse
func convertChunkToDTO(chunk *ai.StreamChunk) *v1.ChatCompletionResponse {
	choices := make([]v1.ChatChoice, len(chunk.Choices))
	for i, ch := range chunk.Choices {
		choice := v1.ChatChoice{
			Index: ch.Index,
		}
		if ch.Delta != nil {
			choice.Delta = &v1.ChatMessage{
				Role:    ch.Delta.Role,
				Content: ch.Delta.Content,
			}
		}
		choices[i] = choice
	}

	return &v1.ChatCompletionResponse{
		ID:      chunk.ID,
		Object:  chunk.Object,
		Created: chunk.Created,
		Model:   chunk.Model,
		Choices: choices,
	}
}
