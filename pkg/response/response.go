package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorBody API 错误体，符合设计文档规范
type ErrorBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// JSON 统一成功 JSON 响应
func JSON(c *gin.Context, status int, data interface{}) {
	c.JSON(status, data)
}

// OK 200 + data
func OK(c *gin.Context, data interface{}) {
	JSON(c, http.StatusOK, data)
}

// Created 201 + data
func Created(c *gin.Context, data interface{}) {
	JSON(c, http.StatusCreated, data)
}

// NoContent 204
func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// Error 统一错误响应
func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorBody{
		Error: struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}{Code: code, Message: message},
	})
}
