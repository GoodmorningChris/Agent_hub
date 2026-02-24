package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"agent-hub/internal/interaction/service"
	"agent-hub/internal/middleware"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
)

// VoteHandler 投票 HTTP 接口
type VoteHandler struct {
	interactionService *service.InteractionService
}

// NewVoteHandler 创建投票 Handler
func NewVoteHandler(interactionService *service.InteractionService) *VoteHandler {
	return &VoteHandler{interactionService: interactionService}
}

// PostVote POST /api/v1/posts/:post_id/vote
func (h *VoteHandler) PostVote(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	postID, err := strconv.ParseInt(c.Param("post_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid post_id")
		return
	}

	var body struct {
		VoteType int `json:"vote_type" binding:"required"` // 1: upvote, -1: downvote
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	netVotes, err := h.interactionService.VotePost(c.Request.Context(), agentID, postID, int8(body.VoteType))
	if err != nil {
		switch err {
		case service.ErrPostNotFound:
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Post not found")
			return
		default:
			response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
			return
		}
	}

	response.OK(c, gin.H{"net_votes": netVotes})
}

// CommentVote POST /api/v1/comments/:comment_id/vote
func (h *VoteHandler) CommentVote(c *gin.Context) {
	agentID := middleware.MustGetAgentID(c)
	commentID, err := strconv.ParseInt(c.Param("comment_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "invalid comment_id")
		return
	}

	var body struct {
		VoteType int `json:"vote_type" binding:"required"` // 1: upvote, -1: downvote
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
		return
	}

	netVotes, err := h.interactionService.VoteComment(c.Request.Context(), agentID, commentID, int8(body.VoteType))
	if err != nil {
		switch err {
		case service.ErrCommentNotFound:
			response.Error(c, http.StatusNotFound, pkgerrors.CodeNotFound, "Comment not found")
			return
		default:
			response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, err.Error())
			return
		}
	}

	response.OK(c, gin.H{"net_votes": netVotes})
}
