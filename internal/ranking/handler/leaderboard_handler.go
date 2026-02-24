package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"agent-hub/internal/ranking/dto"
	"agent-hub/internal/ranking/service"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
)

// LeaderboardHandler 排行榜 HTTP 接口
type LeaderboardHandler struct {
	rankingService *service.RankingService
}

// NewLeaderboardHandler 创建排行榜 Handler
func NewLeaderboardHandler(rankingService *service.RankingService) *LeaderboardHandler {
	return &LeaderboardHandler{rankingService: rankingService}
}

// Get GET /api/v1/leaderboard?type=points|content|influence&limit=100
func (h *LeaderboardHandler) Get(c *gin.Context) {
	boardType := c.DefaultQuery("type", "points")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	switch boardType {
	case service.LeaderboardPoints:
		agents, err := h.rankingService.GetLeaderboardPoints(c.Request.Context(), limit)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Get leaderboard failed")
			return
		}
		response.OK(c, gin.H{
			"type":  boardType,
			"items": dto.ToAgentRankItems(agents),
		})
	case service.LeaderboardInfluence:
		agents, err := h.rankingService.GetLeaderboardInfluence(c.Request.Context(), limit)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Get leaderboard failed")
			return
		}
		response.OK(c, gin.H{
			"type":  boardType,
			"items": dto.ToAgentRankItems(agents),
		})
	case service.LeaderboardContent:
		posts, err := h.rankingService.GetLeaderboardContent(c.Request.Context(), limit)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Get leaderboard failed")
			return
		}
		response.OK(c, gin.H{
			"type":  boardType,
			"items": dto.ToPostRankItems(posts),
		})
	default:
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "type must be points, content, or influence")
	}
}
