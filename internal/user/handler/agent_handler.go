package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
	"agent-hub/internal/user/dto"
	"agent-hub/internal/user/service"
	"agent-hub/internal/middleware"
)

// AgentHandler Agent 相关 HTTP 接口
type AgentHandler struct {
	userService    *service.UserService
	jwtSecret      []byte
	jwtExpireHours int
}

// NewAgentHandler 创建 Agent Handler
func NewAgentHandler(userService *service.UserService, jwtSecret []byte, jwtExpireHours int) *AgentHandler {
	return &AgentHandler{
		userService:    userService,
		jwtSecret:     jwtSecret,
		jwtExpireHours: jwtExpireHours,
	}
}

// Create 创建 Agent POST /api/v1/agents（需认证）
func (h *AgentHandler) Create(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var in service.CreateAgentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	a, token, err := h.userService.CreateAgent(c.Request.Context(), userID, in, h.jwtSecret, h.jwtExpireHours)
	if err != nil {
		switch err {
		case service.ErrAgentExists:
			response.Error(c, http.StatusConflict, pkgerrors.CodeConflict, "You already have an agent")
			return
		case service.ErrAgentNameTaken:
			response.Error(c, http.StatusConflict, pkgerrors.CodeConflict, "Agent name already taken")
			return
		default:
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Create agent failed")
			return
		}
	}

	response.JSON(c, http.StatusCreated, gin.H{
		"agent": dto.ToAgentPublicResponse(a),
		"token": token,
	})
}

// GetByName 获取 Agent 公开信息 GET /api/v1/agents/:agent_name
func (h *AgentHandler) GetByName(c *gin.Context) {
	agentName := c.Param("agent_name")
	if agentName == "" {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "agent_name is required")
		return
	}

	a, err := h.userService.GetAgentByName(c.Request.Context(), agentName)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Get agent failed")
		return
	}
	if a == nil {
		response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Agent not found")
		return
	}

	response.OK(c, dto.ToAgentPublicResponse(a))
}

// UpdateMe 更新当前用户的 Agent PUT /api/v1/me/agent（需认证）
func (h *AgentHandler) UpdateMe(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var in service.UpdateAgentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	a, err := h.userService.UpdateAgent(c.Request.Context(), userID, in)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Update agent failed")
		return
	}
	if a == nil {
		response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Agent not found")
		return
	}

	response.OK(c, dto.ToAgentPublicResponse(a))
}
