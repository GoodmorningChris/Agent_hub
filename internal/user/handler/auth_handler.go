package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"agent-hub/internal/user/service"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
)

// AuthHandler 认证相关 HTTP 接口（注册、登录、OAuth）
type AuthHandler struct {
	userService    *service.UserService
	jwtSecret      []byte
	jwtExpireHours int
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(userService *service.UserService, jwtSecret []byte, jwtExpireHours int) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		jwtSecret:     jwtSecret,
		jwtExpireHours: jwtExpireHours,
	}
}

// Register 用户注册 POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var in service.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	out, err := h.userService.Register(c.Request.Context(), in, h.jwtSecret, h.jwtExpireHours)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			response.Error(c, http.StatusConflict, pkgerrors.CodeConflict, "Email or username already registered")
			return
		default:
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Registration failed")
			return
		}
	}

	response.JSON(c, http.StatusCreated, out)
}

// Login 用户登录 POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var in service.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	token, err := h.userService.Login(c.Request.Context(), in, h.jwtSecret, h.jwtExpireHours)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			response.Error(c, http.StatusUnauthorized, pkgerrors.CodeUnauthorized, "Invalid email or password")
			return
		}
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Login failed")
		return
	}

	response.OK(c, gin.H{"token": token})
}

// OAuthTwitter 跳转 Twitter OAuth GET /api/v1/auth/oauth/twitter
func (h *AuthHandler) OAuthTwitter(c *gin.Context) {
	// TODO: 实现 Twitter OAuth 跳转
	c.Redirect(http.StatusFound, "/")
}

// OAuthTwitterCallback Twitter 回调 GET /api/v1/auth/oauth/twitter/callback
func (h *AuthHandler) OAuthTwitterCallback(c *gin.Context) {
	// TODO: 处理 code，完成绑定
	response.OK(c, gin.H{"message": "OAuth callback not implemented yet"})
}
