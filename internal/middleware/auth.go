package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/jwt"
	"agent-hub/pkg/response"
)

const (
	// ContextKeyUserID  context 中存储的 user_id key
	ContextKeyUserID = "user_id"
	// ContextKeyAgentID context 中存储的 agent_id key
	ContextKeyAgentID = "agent_id"
)

// JWT 解析并校验 JWT，将 user_id、agent_id 写入 context
// 未携带或无效 token 时返回 401
func JWT(secret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || len(auth) < 8 || auth[:7] != "Bearer " {
			response.Error(c, http.StatusUnauthorized, pkgerrors.CodeUnauthorized, "Missing or invalid authorization header")
			c.Abort()
			return
		}
		tokenString := auth[7:]
		claims, err := jwt.Parse(secret, tokenString)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, pkgerrors.CodeUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyAgentID, claims.AgentID)
		c.Next()
	}
}

// GetUserID 从 context 获取当前登录用户的 user_id（需在 JWT 中间件之后调用）
func GetUserID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextKeyUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// GetAgentID 从 context 获取当前登录用户的 agent_id
func GetAgentID(c *gin.Context) (int64, bool) {
	v, ok := c.Get(ContextKeyAgentID)
	if !ok {
		return 0, false
	}
	id, ok := v.(int64)
	return id, ok
}

// MustGetUserID 获取 user_id，若不存在则 panic（仅用于已确认有 JWT 的路由）
func MustGetUserID(c *gin.Context) int64 {
	id, ok := GetUserID(c)
	if !ok {
		panic("user_id not in context")
	}
	return id
}

// MustGetAgentID 获取 agent_id
func MustGetAgentID(c *gin.Context) int64 {
	id, ok := GetAgentID(c)
	if !ok {
		panic("agent_id not in context")
	}
	return id
}
