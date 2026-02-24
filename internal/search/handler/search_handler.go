package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	contentDto "agent-hub/internal/content/dto"
	searchDto "agent-hub/internal/search/dto"
	"agent-hub/internal/search/service"
	userDto "agent-hub/internal/user/dto"
	pkgerrors "agent-hub/pkg/errors"
	"agent-hub/pkg/response"
)

// SearchHandler 搜索 HTTP 接口
type SearchHandler struct {
	searchService *service.SearchService
}

// NewSearchHandler 创建搜索 Handler
func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

// Search GET /api/v1/search?q=xxx&type=posts|agents|all&limit=&offset=
//
// type=agents  仅搜索 Agent
// type=posts   仅搜索帖子
// type=all     同时搜索 Agent 和帖子，结果交错返回（默认）
//
// 支持空格分词：多个关键词用空格分隔，记录须同时匹配所有词（AND 语义）
func (h *SearchHandler) Search(c *gin.Context) {
	q := c.Query("q")
	searchType := c.DefaultQuery("type", service.SearchTypeAll)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	switch searchType {
	case service.SearchTypeAgents:
		agents, total, err := h.searchService.SearchAgents(c.Request.Context(), q, limit, offset)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Search failed")
			return
		}
		items := make([]interface{}, len(agents))
		for i, a := range agents {
			items[i] = userDto.ToAgentPublicResponse(a)
		}
		response.OK(c, gin.H{
			"type":  searchType,
			"query": q,
			"total": total,
			"items": items,
		})

	case service.SearchTypePosts:
		posts, total, err := h.searchService.SearchPosts(c.Request.Context(), q, limit, offset)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Search failed")
			return
		}
		items := make([]interface{}, len(posts))
		for i, p := range posts {
			items[i] = contentDto.ToPostResponse(p)
		}
		response.OK(c, gin.H{
			"type":  searchType,
			"query": q,
			"total": total,
			"items": items,
		})

	case service.SearchTypeAll:
		result, err := h.searchService.SearchAll(c.Request.Context(), q, limit, offset)
		if err != nil {
			response.Error(c, http.StatusInternalServerError, pkgerrors.CodeInternal, "Search failed")
			return
		}

		// 交错拼装结果：agent、post、agent、post …
		items := make([]gin.H, 0, len(result.Agents)+len(result.Posts))
		ai, pi := 0, 0
		for ai < len(result.Agents) || pi < len(result.Posts) {
			if ai < len(result.Agents) {
				items = append(items, gin.H{
					"type": searchDto.ResultTypeAgent,
					"data": userDto.ToAgentPublicResponse(result.Agents[ai]),
				})
				ai++
			}
			if pi < len(result.Posts) {
				items = append(items, gin.H{
					"type": searchDto.ResultTypePost,
					"data": contentDto.ToPostResponse(result.Posts[pi]),
				})
				pi++
			}
		}

		response.OK(c, gin.H{
			"type":         service.SearchTypeAll,
			"query":        result.Query,
			"total_agents": result.TotalAgents,
			"total_posts":  result.TotalPosts,
			"total":        result.TotalAgents + result.TotalPosts,
			"items":        items,
		})

	default:
		response.Error(c, http.StatusBadRequest, pkgerrors.CodeInvalidRequest, "type must be agents, posts or all")
	}
}
