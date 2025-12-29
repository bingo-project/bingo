// ABOUTME: HTTP handlers for notification center endpoints.
// ABOUTME: Provides list, read, and delete operations for notifications.

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

// NotificationHandler handles notification requests.
type NotificationHandler struct {
	notificationBiz notification.NotificationBiz
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(ds store.IStore) *NotificationHandler {
	return &NotificationHandler{
		notificationBiz: notification.New(ds),
	}
}

// List
// @Summary    List notifications
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      category   query     string  false  "Filter by category"
// @Param      is_read    query     bool    false  "Filter by read status"
// @Param      page       query     int     false  "Page number"
// @Param      page_size  query     int     false  "Page size"
// @Success    200        {object}  v1.ListNotificationsResponse
// @Failure    400        {object}  core.ErrResponse
// @Failure    500        {object}  core.ErrResponse
// @Router     /v1/notifications [GET].
func (h *NotificationHandler) List(c *gin.Context) {
	var req v1.ListNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	userID := contextx.UserID(c)
	resp, err := h.notificationBiz.List(c, userID, &req)
	core.Response(c, resp, err)
}

// UnreadCount
// @Summary    Get unread notification count
// @Security   Bearer
// @Tags       Notification
// @Produce    json
// @Success    200  {object}  v1.UnreadCountResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/unread-count [GET].
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	userID := contextx.UserID(c)
	resp, err := h.notificationBiz.UnreadCount(c, userID)
	core.Response(c, resp, err)
}

// MarkAsRead
// @Summary    Mark notification as read
// @Security   Bearer
// @Tags       Notification
// @Param      uuid  path  string  true  "Notification UUID"
// @Success    200
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/{uuid}/read [PUT].
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	uuid := c.Param("uuid")
	userID := contextx.UserID(c)
	err := h.notificationBiz.MarkAsRead(c, userID, uuid)
	core.Response(c, nil, err)
}

// MarkAllAsRead
// @Summary    Mark all notifications as read
// @Security   Bearer
// @Tags       Notification
// @Success    200
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/read-all [PUT].
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := contextx.UserID(c)
	err := h.notificationBiz.MarkAllAsRead(c, userID)
	core.Response(c, nil, err)
}

// Delete
// @Summary    Delete notification
// @Security   Bearer
// @Tags       Notification
// @Param      uuid  path  string  true  "Notification UUID"
// @Success    200
// @Failure    403  {object}  core.ErrResponse
// @Failure    404  {object}  core.ErrResponse
// @Failure    500  {object}  core.ErrResponse
// @Router     /v1/notifications/{uuid} [DELETE].
func (h *NotificationHandler) Delete(c *gin.Context) {
	uuid := c.Param("uuid")
	userID := contextx.UserID(c)
	err := h.notificationBiz.Delete(c, userID, uuid)
	core.Response(c, nil, err)
}
