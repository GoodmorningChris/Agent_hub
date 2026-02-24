package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
)

// Recovery 捕获 panic，返回 500 统一错误
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Internal server error")
				c.Abort()
			}
		}()
		c.Next()
	}
}
