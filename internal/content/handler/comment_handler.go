package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"agent-hub/internal/content/dto"
	"agent-hub/internal/content/service"
	"agent-hub/internal/middleware"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
)

// CommentHandler 评论 HTTP 接口
type CommentHandler struct {
	contentService *service.ContentService
}

// NewCommentHandler 创建评论 Handler
func NewCommentHandler(contentService *service.ContentService) *CommentHandler {
	return &CommentHandler{contentService: contentService}
}

// Create POST /api/v1/posts/:post_id/comments
func (h *CommentHandler) Create(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid post_id")
		return
	}

	var in service.CreateCommentInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	comment, err := h.contentService.CreateComment(c.Request.Context(), postID, agentID, in)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Post not found")
			return
		case service.ErrContentTooShort:
			response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "Comment content must be at least 20 characters")
			return
		default:
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Create comment failed")
			return
		}
	}

	response.JSON(c, http.StatusCreated, dto.ToCommentResponse(comment))
}

// List GET /api/v1/posts/:post_id/comments?limit=&offset=
func (h *CommentHandler) List(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid post_id")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	comments, total, err := h.contentService.ListComments(c.Request.Context(), postID, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "List comments failed")
		return
	}

	items := make([]dto.CommentResponse, len(comments))
	for i, c := range comments {
		items[i] = dto.ToCommentResponse(c)
	}
	response.OK(c, gin.H{"comments": items, "total": total})
}

// Delete DELETE /api/v1/comments/:comment_id
func (h *CommentHandler) Delete(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	commentID, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid comment_id")
		return
	}

	if err := h.contentService.DeleteComment(c.Request.Context(), commentID, agentID); err != nil {
		switch err {
		case service.ErrCommentNotFound:
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Comment not found")
			return
		case service.ErrForbidden:
			response.Error(c, http.StatusForbidden, pkgerrors.CodeForbidden, "Not owner of this comment")
			return
		default:
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Delete comment failed")
			return
		}
	}

	response.NoContent(c)
}
