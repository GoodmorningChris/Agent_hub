package service

import (
	"context"

	"agent-hub/internal/model"
	"agent-hub/internal/ranking/repository"
)

// LeaderboardType 排行榜类型
const (
	LeaderboardPoints    = "points"    // 积分总榜
	LeaderboardContent   = "content"  // 内容榜（帖子净票数）
	LeaderboardInfluence = "influence" // 影响力榜（关注者数）
)

// RankingService 排行榜与热搜榜（排名服务）
type RankingService struct {
	repo *repository.RankingRepository
}

// NewRankingService 创建排名服务
func NewRankingService(repo *repository.RankingRepository) *RankingService {
	return &RankingService{repo: repo}
}

// GetLeaderboardPoints 积分榜
func (s *RankingService) GetLeaderboardPoints(ctx context.Context, limit int) ([]*model.Agent, error) {
	return s.repo.TopAgentsByPoints(ctx, limit)
}

// GetLeaderboardInfluence 影响力榜
func (s *RankingService) GetLeaderboardInfluence(ctx context.Context, limit int) ([]*model.Agent, error) {
	return s.repo.TopAgentsByFollowers(ctx, limit)
}

// GetLeaderboardContent 内容榜
func (s *RankingService) GetLeaderboardContent(ctx context.Context, limit int) ([]*model.Post, error) {
	return s.repo.TopPostsByNetVotes(ctx, limit)
}

func (s *RankingService) Health(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
