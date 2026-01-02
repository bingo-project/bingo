package ai

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/bingo-project/bingo/internal/pkg/model"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
)

func TestAiAgentBiz_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)

		// Create the table manually to avoid SQLite migration issues
		err = db.Exec(`
			CREATE TABLE ai_agents (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				agent_id TEXT(32) NOT NULL UNIQUE,
				name TEXT(64) NOT NULL,
				description TEXT(255),
				icon TEXT(255),
				category TEXT(32) NOT NULL DEFAULT 'general',
				system_prompt TEXT NOT NULL,
				model TEXT(64),
				temperature REAL DEFAULT 0.7,
				max_tokens INTEGER DEFAULT 2000,
				sort INTEGER DEFAULT 0,
				status TEXT(16) NOT NULL DEFAULT 'active',
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`).Error
		require.NoError(t, err)

		ds := store.NewStore(db)
		biz := NewAiAgent(ds)

		req := &v1.CreateAiAgentRequest{
			AgentID:      "test-agent",
			Name:         "Test Agent",
			Description:  "Test description",
			SystemPrompt: "You are a test agent",
			Model:        "gpt-4",
		}

		resp, err := biz.Create(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, "test-agent", resp.AgentID)
		assert.Equal(t, "Test Agent", resp.Name)
		assert.Equal(t, string(model.AiAgentStatusActive), resp.Status)
	})
}
