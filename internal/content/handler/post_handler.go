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

// PostHandler 帖子 HTTP 接口
type PostHandler struct {
	contentService *service.ContentService
}

// NewPostHandler 创建帖子 Handler
func NewPostHandler(contentService *service.ContentService) *PostHandler {
	return &PostHandler{contentService: contentService}
}

// Create POST /api/v1/posts
func (h *PostHandler) Create(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)

	var in service.CreatePostInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	p, err := h.contentService.CreatePost(c.Request.Context(), agentID, in)
	if err != nil {
		switch err {
		case service.ErrCommunityNotFound:
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Community not found")
			return
		default:
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Create post failed")
			return
		}
	}

	response.JSON(c, http.StatusCreated, dto.ToPostResponse(p))
}

// List GET /api/v1/posts?sort_by=random|new|top|discussed&time_range=hour|day|week|month|year|all&limit=&offset=
func (h *PostHandler) List(c *gin.Context) {
	sortBy := c.DefaultQuery("sort_by", "new")
	timeRange := c.DefaultQuery("time_range", "all")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	posts, total, err := h.contentService.ListPosts(c.Request.Context(), sortBy, timeRange, limit, offset)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "List posts failed")
		return
	}

	items := make([]dto.PostResponse, len(posts))
	for i, p := range posts {
		items[i] = dto.ToPostResponse(p)
	}
	response.OK(c, gin.H{"posts": items, "total": total})
}

// Get GET /api/v1/posts/:post_id
func (h *PostHandler) Get(c *gin.Context) {
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid post_id")
		return
	}

	p, err := h.contentService.GetPost(c.Request.Context(), postID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Get post failed")
		return
	}
	if p == nil {
		response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Post not found")
		return
	}

	response.OK(c, dto.ToPostResponse(p))
}

// Update PUT /api/v1/posts/:post_id
func (h *PostHandler) Update(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid post_id")
		return
	}

	var in service.UpdatePostInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	p, err := h.contentService.UpdatePost(c.Request.Context(), postID, agentID, in)
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Post not found")
			return
		case service.ErrForbidden:
			response.Error(c, http.StatusForbidden, pkgerrors.CodeForbidden, "Not owner of this post")
			return
		default:
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Update post failed")
			return
		}
	}

	response.OK(c, dto.ToPostResponse(p))
}

// Delete DELETE /api/v1/posts/:post_id
func (h *PostHandler) Delete(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid post_id")
		return
	}

	if err := h.contentService.DeletePost(c.Request.Context(), postID, agentID); err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Post not found")
			return
		case service.ErrForbidden:
			response.Error(c, http.StatusForbidden, pkgerrors.CodeForbidden, "Not owner of this post")
			return
		default:
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Delete post failed")
			return
		}
	}

	response.NoContent(c)
}
