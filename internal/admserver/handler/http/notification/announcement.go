// ABOUTME: HTTP handlers for announcement management.
// ABOUTME: Provides CRUD, publish, schedule, and cancel operations for announcements.

package notification

import (
	"github.com/gin-gonic/gin"

	"github.com/bingo-project/bingo/internal/admserver/biz"
	"github.com/bingo-project/bingo/internal/pkg/auth"
	"github.com/bingo-project/bingo/internal/pkg/core"
	"github.com/bingo-project/bingo/internal/pkg/errno"
	"github.com/bingo-project/bingo/internal/pkg/log"
	"github.com/bingo-project/bingo/internal/pkg/store"
	v1 "github.com/bingo-project/bingo/pkg/api/admserver/v1"
)

type AnnouncementHandler struct {
	a *auth.Authorizer
	b biz.IBiz
}

func NewAnnouncementHandler(ds store.IStore, a *auth.Authorizer) *AnnouncementHandler {
	return &AnnouncementHandler{a: a, b: biz.NewBiz(ds)}
}

// List
// @Summary    List announcements
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      request	 query	    v1.ListAnnouncementsRequest	 true  "Param"
// @Success	   200		{object}	v1.ListAnnouncementsResponse
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/announcements [GET].
func (h *AnnouncementHandler) List(c *gin.Context) {
	log.C(c).Infow("List announcements function called")

	var req v1.ListAnnouncementsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := h.b.Announcements().List(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Get
// @Summary    Get announcement info
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      uuid	     path	    string            		 true  "UUID"
// @Success	   200		{object}	v1.AnnouncementItem
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/announcements/{uuid} [GET].
func (h *AnnouncementHandler) Get(c *gin.Context) {
	log.C(c).Infow("Get announcement function called")

	uuid := c.Param("uuid")
	resp, err := h.b.Announcements().Get(c, uuid)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Create
// @Summary    Create announcement
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      request	 body	    v1.CreateAnnouncementRequest	 true  "Param"
// @Success	   200		{object}	v1.AnnouncementItem
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/announcements [POST].
func (h *AnnouncementHandler) Create(c *gin.Context) {
	log.C(c).Infow("Create announcement function called")

	var req v1.CreateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	resp, err := h.b.Announcements().Create(c, &req)
	if err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, resp, nil)
}

// Update
// @Summary    Update announcement info
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      uuid	     path	    string            		 true  "UUID"
// @Param      request	 body	    v1.UpdateAnnouncementRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/announcements/{uuid} [PUT].
func (h *AnnouncementHandler) Update(c *gin.Context) {
	log.C(c).Infow("Update announcement function called")

	var req v1.UpdateAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	uuid := c.Param("uuid")
	if err := h.b.Announcements().Update(c, uuid, &req); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// Delete
// @Summary    Delete announcement
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      uuid	    path	    string            true  "UUID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/announcements/{uuid} [DELETE].
func (h *AnnouncementHandler) Delete(c *gin.Context) {
	log.C(c).Infow("Delete announcement function called")

	uuid := c.Param("uuid")
	if err := h.b.Announcements().Delete(c, uuid); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// Publish
// @Summary    Publish announcement immediately
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      uuid	    path	    string            true  "UUID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/announcements/{uuid}/publish [POST].
func (h *AnnouncementHandler) Publish(c *gin.Context) {
	log.C(c).Infow("Publish announcement function called")

	uuid := c.Param("uuid")
	if err := h.b.Announcements().Publish(c, uuid); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// Schedule
// @Summary    Schedule announcement for later publication
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      uuid	    path	    string            true  "UUID"
// @Param      request	 body	    v1.ScheduleAnnouncementRequest	 true  "Param"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/announcements/{uuid}/schedule [POST].
func (h *AnnouncementHandler) Schedule(c *gin.Context) {
	log.C(c).Infow("Schedule announcement function called")

	var req v1.ScheduleAnnouncementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		core.Response(c, nil, errno.ErrInvalidArgument.WithMessage("%s", err.Error()))

		return
	}

	uuid := c.Param("uuid")
	if err := h.b.Announcements().Schedule(c, uuid, &req); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}

// Cancel
// @Summary    Cancel scheduled announcement
// @Security   Bearer
// @Tags       Notification
// @Accept     application/json
// @Produce    json
// @Param      uuid	    path	    string            true  "UUID"
// @Success	   200		{object}	nil
// @Failure	   400		{object}	core.ErrResponse
// @Failure	   500		{object}	core.ErrResponse
// @Router    /v1/announcements/{uuid}/cancel [POST].
func (h *AnnouncementHandler) Cancel(c *gin.Context) {
	log.C(c).Infow("Cancel announcement function called")

	uuid := c.Param("uuid")
	if err := h.b.Announcements().Cancel(c, uuid); err != nil {
		core.Response(c, nil, err)

		return
	}

	core.Response(c, nil, nil)
}
