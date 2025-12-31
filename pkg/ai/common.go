// ABOUTME: Common utilities for AI providers.
// ABOUTME: Shared conversion and helper functions used by all providers.

package ai

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/cloudwego/eino/schema"
)

// GenerateID generates a unique ID for chat completions.
func GenerateID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "chatcmpl-" + hex.EncodeToString(b)
}

// ConvertMessages converts ai.Message to schema.Message.
func ConvertMessages(msgs []Message) []*schema.Message {
	result := make([]*schema.Message, len(msgs))
	for i, m := range msgs {
		role := schema.User
		switch m.Role {
		case RoleSystem:
			role = schema.System
		case RoleAssistant:
			role = schema.Assistant
		}
		result[i] = &schema.Message{
			Role:    role,
			Content: m.Content,
		}
	}
	return result
}

// ConvertResponse converts Eino response to ai.ChatResponse.
func ConvertResponse(resp *schema.Message, modelName string) *ChatResponse {
	usage := ExtractUsage(resp)

	return &ChatResponse{
		ID:      GenerateID(),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []Choice{
			{
				Index: 0,
				Message: Message{
					Role:    RoleAssistant,
					Content: resp.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: usage,
	}
}

// ConvertStreamChunk converts Eino stream message to ai.StreamChunk.
func ConvertStreamChunk(msg *schema.Message, modelName string, id string) *StreamChunk {
	chunk := &StreamChunk{
		ID:      id,
		Object:  "chat.completion.chunk",
		Created: time.Now().Unix(),
		Model:   modelName,
		Choices: []Choice{
			{
				Index: 0,
				Delta: &Message{
					Role:    RoleAssistant,
					Content: msg.Content,
				},
			},
		},
	}

	// Extract usage if present (typically in the last chunk)
	if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
		chunk.Usage = &Usage{
			PromptTokens:     msg.ResponseMeta.Usage.PromptTokens,
			CompletionTokens: msg.ResponseMeta.Usage.CompletionTokens,
			TotalTokens:      msg.ResponseMeta.Usage.TotalTokens,
		}
	}

	return chunk
}

// ExtractUsage extracts token usage from Eino message.
func ExtractUsage(msg *schema.Message) Usage {
	if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
		return Usage{
			PromptTokens:     msg.ResponseMeta.Usage.PromptTokens,
			CompletionTokens: msg.ResponseMeta.Usage.CompletionTokens,
			TotalTokens:      msg.ResponseMeta.Usage.TotalTokens,
		}
	}
	return Usage{}
}
