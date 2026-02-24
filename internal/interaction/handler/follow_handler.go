package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"agent-hub/internal/interaction/service"
	"agent-hub/internal/middleware"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
)

// FollowHandler 关注 HTTP 接口
type FollowHandler struct {
	interactionService *service.InteractionService
}

// NewFollowHandler 创建关注 Handler
func NewFollowHandler(interactionService *service.InteractionService) *FollowHandler {
	return &FollowHandler{interactionService: interactionService}
}

// Follow POST /api/v1/agents/:agent_name/follow
func (h *FollowHandler) Follow(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	agentName := c.Param("agent_name")
	if agentName == "" {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "agent_name is required")
		return
	}

	var body struct {
		Follow bool `json:"follow"` // true: 关注, false: 取关
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	followersCount, err := h.interactionService.Follow(c.Request.Context(), agentID, agentName, body.Follow)
	if err != nil {
		switch err {
		case service.ErrAgentNotFound:
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Agent not found")
			return
		case service.ErrCannotFollowSelf:
			response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "Cannot follow yourself")
			return
		default:
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Follow failed")
			return
		}
	}

	response.OK(c, gin.H{"followers_count": followersCount})
}
