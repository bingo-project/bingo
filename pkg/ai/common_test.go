// ABOUTME: Unit tests for common AI provider utilities.
// ABOUTME: Tests for ID generation, message conversion, and usage extraction.

package ai

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestGenerateID(t *testing.T) {
	id := GenerateID()
	if len(id) == 0 {
		t.Fatal("GenerateID returned empty string")
	}
	if len(id) < 9 {
		t.Fatalf("GenerateID returned too short string: %s", id)
	}
	if id[:9] != "chatcmpl-" {
		t.Fatalf("GenerateID has wrong prefix: %s", id[:9])
	}

	// Test that IDs are unique
	id2 := GenerateID()
	if id == id2 {
		t.Fatal("GenerateID returned duplicate IDs")
	}
}

func TestConvertMessages(t *testing.T) {
	msgs := []Message{
		{Role: RoleSystem, Content: "You are helpful"},
		{Role: RoleUser, Content: "Hello"},
		{Role: RoleAssistant, Content: "Hi there"},
	}

	converted := ConvertMessages(msgs)

	if len(converted) != 3 {
		t.Fatalf("Expected 3 messages, got %d", len(converted))
	}

	if converted[0].Role != schema.System {
		t.Errorf("Expected System role, got %v", converted[0].Role)
	}
	if converted[1].Role != schema.User {
		t.Errorf("Expected User role, got %v", converted[1].Role)
	}
	if converted[2].Role != schema.Assistant {
		t.Errorf("Expected Assistant role, got %v", converted[2].Role)
	}

	if converted[0].Content != "You are helpful" {
		t.Errorf("Expected 'You are helpful', got '%s'", converted[0].Content)
	}
}

func TestConvertMessagesEmpty(t *testing.T) {
	msgs := []Message{}
	converted := ConvertMessages(msgs)

	if len(converted) != 0 {
		t.Fatalf("Expected 0 messages, got %d", len(converted))
	}
}

func TestExtractUsage(t *testing.T) {
	msg := &schema.Message{
		ResponseMeta: &schema.ResponseMeta{
			Usage: &schema.TokenUsage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		},
	}

	usage := ExtractUsage(msg)
	if usage.PromptTokens != 10 {
		t.Errorf("Expected 10 prompt tokens, got %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 20 {
		t.Errorf("Expected 20 completion tokens, got %d", usage.CompletionTokens)
	}
	if usage.TotalTokens != 30 {
		t.Errorf("Expected 30 total tokens, got %d", usage.TotalTokens)
	}
}

func TestExtractUsageNil(t *testing.T) {
	msg := &schema.Message{}

	usage := ExtractUsage(msg)
	if usage.PromptTokens != 0 {
		t.Errorf("Expected 0 prompt tokens, got %d", usage.PromptTokens)
	}
	if usage.CompletionTokens != 0 {
		t.Errorf("Expected 0 completion tokens, got %d", usage.CompletionTokens)
	}
	if usage.TotalTokens != 0 {
		t.Errorf("Expected 0 total tokens, got %d", usage.TotalTokens)
	}
}

func TestConvertResponse(t *testing.T) {
	msg := &schema.Message{
		Content: "Hello world",
		ResponseMeta: &schema.ResponseMeta{
			Usage: &schema.TokenUsage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
		},
	}

	resp := ConvertResponse(msg, "gpt-4")

	if resp.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", resp.Model)
	}
	if resp.ID == "" {
		t.Error("Expected non-empty ID")
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("Expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello world" {
		t.Errorf("Expected 'Hello world', got '%s'", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 30 {
		t.Errorf("Expected 30 total tokens, got %d", resp.Usage.TotalTokens)
	}
}

func TestConvertStreamChunk(t *testing.T) {
	msg := &schema.Message{
		Content: "Hello",
		ResponseMeta: &schema.ResponseMeta{
			Usage: &schema.TokenUsage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		},
	}

	chunk := ConvertStreamChunk(msg, "gpt-4", "test-id")

	if chunk.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got '%s'", chunk.ID)
	}
	if chunk.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", chunk.Model)
	}
	if chunk.Choices[0].Delta.Content != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", chunk.Choices[0].Delta.Content)
	}
	if chunk.Usage.TotalTokens != 15 {
		t.Errorf("Expected 15 total tokens, got %d", chunk.Usage.TotalTokens)
	}
}
