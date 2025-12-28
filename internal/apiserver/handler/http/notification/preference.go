// ABOUTME: HTTP handlers for notification preference endpoints.
// ABOUTME: Provides get and update operations for user notification settings.

package notification

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/apiserver/biz/notification"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/apiserver/v1"
	"github.com/bingo-project/bingo/pkg/contextx"
)

// PreferenceHandler handles notification preference requests.
type PreferenceHandler struct {
	preferenceBiz notification.PreferenceBiz
}

// NewPreferenceHandler creates a new PreferenceHandler.
func NewPreferenceHandler(ds store.IStore) *PreferenceHandler {
	return &PreferenceHandler{
		preferenceBiz: notification.NewPreference(ds),
	}
}

// Get
// @Summary    Get notification preferences
// @Security   Bearer
// @Tags       Notification
// @Produce    json
// @Success    200  {object}  v1.NotificationPreferences
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/preferences [GET].
func (h *PreferenceHandler) Get(c *gin.Context) {
	userID := contextx.UserID(c)
	resp, err := h.preferenceBiz.Get(c, userID)
	core.Response(c, resp, err)
}

// Update
// @Summary    Update notification preferences
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      request  body  v1.UpdatePreferencesRequest  true  "Preferences"
// @Success    200
// @Failure    400  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/preferences [PUT].
func (h *PreferenceHandler) Update(c *gin.Context) {
	var req v1.UpdatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage(err.Error()))

		return
	}

	userID := contextx.UserID(c)
	err := h.preferenceBiz.Update(c, userID, &req)
	core.Response(c, nil, err)
}
