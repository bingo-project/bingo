// ABOUTME: Session HTTP handlers for AI chat sessions.
// ABOUTME: Provides endpoints for session CRUD and history.

package chat

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"

	"github.com/bingo-project/bingo/internal/apiserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	"github.com/bingo-project/bingo/pkg/ai"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
)

type SessionHandler struct {
	b biz.IBiz
}

func NewSessionHandler(ds store.IStore, registry *ai.Registry) *SessionHandler {
	return &SessionHandler{
		b: biz.NewBiz(ds).WithRegistry(registry),
	}
}

// CreateSession
// @Summary    Create chat session
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      request  body      v1.CreateSessionRequest  true  "Session request"
// @Success    200      {object}  v1.SessionInfo
// @Failure    400      {object}  core.ErrResponse
// @Failure    500      {object}  core.ErrResponse
// @Router     /v1/ai/sessions [POST].
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req v1.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	uid := contextx.UserID(c)
	session, err := h.b.Chat().Sessions().Create(c, uid, req.Title, req.Model)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	var resp v1.SessionInfo
	_ = copier.Copy(&resp, session)
	resp.SessionID = session.SessionID

	core.Response(c, resp, nil)
}

// ListSessions
// @Summary    List chat sessions
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Success    200  {object}  v1.ListSessionsResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/ai/sessions [GET].
func (h *SessionHandler) ListSessions(c *gin.Context) {
	uid := contextx.UserID(c)
	sessions, err := h.b.Chat().Sessions().List(c, uid)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	data := make([]v1.SessionInfo, len(sessions))
	for i, s := range sessions {
		_ = copier.Copy(&data[i], s)
		data[i].SessionID = s.SessionID
	}

	core.Response(c, v1.ListSessionsResponse{Data: data}, nil)
}

// GetSession
// @Summary    Get chat session
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      session_id  path      string  true  "Session ID"
// @Success    200         {object}  v1.SessionInfo
// @Failure    404         {object}  core.ErrResponse
// @Failure    500         {object}  core.ErrResponse
// @Router     /v1/ai/sessions/{session_id} [GET].
func (h *SessionHandler) GetSession(c *gin.Context) {
	uid := contextx.UserID(c)
	sessionID := c.Param("session_id")

	session, err := h.b.Chat().Sessions().Get(c, uid, sessionID)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	var resp v1.SessionInfo
	_ = copier.Copy(&resp, session)
	resp.SessionID = session.SessionID

	core.Response(c, resp, nil)
}

// DeleteSession
// @Summary    Delete chat session
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      session_id  path  string  true  "Session ID"
// @Success    200         {object}  nil
// @Failure    404         {object}  core.ErrResponse
// @Failure    500         {object}  core.ErrResponse
// @Router     /v1/ai/sessions/{session_id} [DELETE].
func (h *SessionHandler) DeleteSession(c *gin.Context) {
	uid := contextx.UserID(c)
	sessionID := c.Param("session_id")

	err := h.b.Chat().Sessions().Delete(c, uid, sessionID)
	core.Response(c, nil, err)
}

// GetSessionHistory
// @Summary    Get session message history
// @Security   Bearer
// @Tags       AI
// @Accept     application/json
// @Produce    json
// @Param      session_id  path      string  true  "Session ID"
// @Param      limit       query     int     false "Limit"
// @Success    200         {object}  v1.SessionHistoryResponse
// @Failure    404         {object}  core.ErrResponse
// @Failure    500         {object}  core.ErrResponse
// @Router     /v1/ai/sessions/{session_id}/history [GET].
func (h *SessionHandler) GetSessionHistory(c *gin.Context) {
	uid := contextx.UserID(c)
	sessionID := c.Param("session_id")

	// Verify session ownership
	_, err := h.b.Chat().Sessions().Get(c, uid, sessionID)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	limit := 100 // default
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	messages, err := h.b.Chat().Sessions().GetHistory(c, sessionID, limit)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	data := make([]v1.ChatMessage, len(messages))
	for i, m := range messages {
		data[i] = v1.ChatMessage{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	core.Response(c, v1.SessionHistoryResponse{
		SessionID: sessionID,
		Messages:  data,
	}, nil)
}
