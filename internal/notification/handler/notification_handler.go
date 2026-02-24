package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"agent-hub/internal/middleware"
	"agent-hub/internal/notification/dto"
	"agent-hub/internal/notification/service"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
)

// NotificationHandler 通知 HTTP 接口
type NotificationHandler struct {
	notificationService *service.NotificationService
}

// NewNotificationHandler 创建通知 Handler
func NewNotificationHandler(notificationService *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// List GET /api/v1/notifications?limit=&offset=
func (h *NotificationHandler) List(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	list, total, err := h.notificationService.List(c.Request.Context(), agentID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "List notifications failed")
		return
	}

	items := make([]dto.NotificationResponse, len(list))
	for i, n := range list {
		items[i] = dto.NotificationResponse{
			ID:                n.ID,
			Type:              n.Type,
			Title:             n.Title,
			Content:           n.Content,
			RelatedEntityID:   n.RelatedEntityID,
			RelatedEntityType: n.RelatedEntityType,
			ActorAgentID:      n.ActorAgentID,
			IsRead:            n.IsRead,
			CreatedAt:         n.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	response.OK(c, gin.H{"notifications": items, "total": total})
}

// MarkRead PATCH /api/v1/notifications/:id/read
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid notification id")
		return
	}

	if err := h.notificationService.MarkRead(c.Request.Context(), id, agentID); err != nil {
		if err == service.ErrNotificationNotFound {
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Notification not found")
			return
		}
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Mark read failed")
		return
	}
	response.NoContent(c)
}

// MarkAllRead POST /api/v1/notifications/read-all
func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)

	if err := h.notificationService.MarkAllRead(c.Request.Context(), agentID); err != nil {
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Mark all read failed")
		return
	}
	response.NoContent(c)
}
