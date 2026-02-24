package service

import (
	"context"

	"agent-hub/internal/model"
	"agent-hub/internal/points/repository"
)

// Adder 供其他模块调用的积分增加接口（避免循环依赖）
type Adder interface {
	AddPoints(ctx context.Context, agentID int64, reason string, relatedEntityID *int64) error
}

// 每日上限：0 表示按「一次性」处理
const (
	DailyCapPostCreated    = 50  // 50 分/日
	DailyCapCommentCreated = 50
	DailyCapContentUpvoted  = 100
	DailyCapDailyLogin     = 5
)

// PointsService 积分计算与管理（积分服务）
type PointsService struct {
	repo *repository.PointsRepository
}

// NewPointsService 创建积分服务
func NewPointsService(repo *repository.PointsRepository) *PointsService {
	return &PointsService{repo: repo}
}

// AddPoints 根据原因增加/扣减积分，并写日志；内部做每日上限与一次性校验
func (s *PointsService) AddPoints(ctx context.Context, agentID int64, reason string, relatedEntityID *int64) error {
	points, oneTime, dailyCap := s.rule(reason)
	if points == 0 {
		return nil
	}

	if oneTime {
		has, err := s.repo.HasReasonOnce(ctx, agentID, reason)
		if err != nil {
			return err
		}
		if has {
			return nil
		}
	} else if dailyCap > 0 && points > 0 {
		sum, err := s.repo.SumTodayPointsByAgentAndReason(ctx, agentID, reason)
		if err != nil {
			return err
		}
		if sum >= dailyCap {
			return nil
		}
		if sum+points > dailyCap {
			points = dailyCap - sum
		}
	}

	if points == 0 {
		return nil
	}

	if err := s.repo.AddAgentPoints(ctx, agentID, points); err != nil {
		return err
	}
	return s.repo.CreateLog(ctx, &model.PointsLog{
		AgentID:         agentID,
		PointsChange:    points,
		Reason:          reason,
		RelatedEntityID: relatedEntityID,
	})
}

// rule 返回 (积分变动, 是否一次性, 每日上限，正数才检查)
func (s *PointsService) rule(reason string) (points int, oneTime bool, dailyCap int) {
	switch reason {
	case model.PointsReasonAgentRegistered:
		return 100, true, 0
	case model.PointsReasonProfileCompleted:
		return 50, true, 0
	case model.PointsReasonPostCreated:
		return 10, false, DailyCapPostCreated
	case model.PointsReasonCommentCreated:
		return 5, false, DailyCapCommentCreated
	case model.PointsReasonContentUpvoted:
		return 1, false, DailyCapContentUpvoted
	case model.PointsReasonDailyLogin:
		return 5, false, DailyCapDailyLogin
	case model.PointsReasonContentDownvoted:
		return -1, false, 0
	case model.PointsReasonContentDeletedByAdmin:
		return -20, false, 0
	default:
		return 0, false, 0
	}
}

func (s *PointsService) Health(ctx context.Context) error {
	return s.repo.Ping(ctx)
}
