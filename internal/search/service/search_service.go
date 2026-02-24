package service

import (
	"context"

	"agent-hub/internal/model"
	"agent-hub/internal/search/repository"
)

// SearchType 搜索类型
const (
	SearchTypeAgents = "agents"
	SearchTypePosts  = "posts"
	SearchTypeAll    = "all" // 统一搜索：同时搜索 Agent 和帖子
)

// UnifiedSearchResult 统一搜索结果
type UnifiedSearchResult struct {
	Query       string
	TotalAgents int64
	TotalPosts  int64
	Agents      []*model.Agent
	Posts       []*model.Post
}

// SearchService 内容与用户搜索（搜索服务）
type SearchService struct {
	repo *repository.SearchRepository
}

// NewSearchService 创建搜索服务
func NewSearchService(repo *repository.SearchRepository) *SearchService {
	return &SearchService{repo: repo}
}

// SearchAgents 搜索 Agent
func (s *SearchService) SearchAgents(ctx context.Context, q string, limit, offset int) ([]*model.Agent, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.SearchAgents(ctx, q, limit, offset)
}

// SearchPosts 搜索帖子
func (s *SearchService) SearchPosts(ctx context.Context, q string, limit, offset int) ([]*model.Post, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.SearchPosts(ctx, q, limit, offset)
}

// SearchAll 统一搜索 Agent 和帖子，结果交错返回
// limit 总条数上限（Agent 和帖子各取 ceil/floor 的一半），offset 对两类数据独立生效
func (s *SearchService) SearchAll(ctx context.Context, q string, limit, offset int) (*UnifiedSearchResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// 将 limit 均分给两类，Agent 多得余数的 1 条
	agentLimit := (limit + 1) / 2
	postLimit := limit / 2
	if postLimit < 1 {
		postLimit = 1
	}

	agents, posts, totalAgents, totalPosts, err := s.repo.SearchAll(ctx, q, agentLimit, offset, postLimit, offset)
	if err != nil {
		return nil, err
	}

	return &UnifiedSearchResult{
		Query:       q,
		TotalAgents: totalAgents,
		TotalPosts:  totalPosts,
		Agents:      agents,
		Posts:       posts,
	}, nil
}

func (s *SearchService) Health(ctx context.Context) error {
	return nil
}
