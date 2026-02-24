package service

import (
	"context"
	"errors"

	"agent-hub/internal/model"
	contentRepo "agent-hub/internal/content/repository"
	"agent-hub/internal/interaction/repository"
	notificationService "agent-hub/internal/notification/service"
	pointsService "agent-hub/internal/points/service"
	userRepo "agent-hub/internal/user/repository"
)

var (
	ErrPostNotFound    = errors.New("post not found")
	ErrCommentNotFound = errors.New("comment not found")
	ErrAgentNotFound   = errors.New("agent not found")
	ErrCannotFollowSelf = errors.New("cannot follow yourself")
)

// InteractionService 投票与关注业务逻辑层（互动服务）
type InteractionService struct {
	voteRepo    *repository.VoteRepository
	followRepo  *repository.FollowRepository
	postRepo    *contentRepo.PostRepository
	commentRepo *contentRepo.CommentRepository
	agentRepo   *userRepo.AgentRepository
	pointsAdder pointsService.Adder
	notifier    notificationService.Notifier
}

// NewInteractionService 创建互动服务，pointsAdder/notifier 可为 nil
func NewInteractionService(
	voteRepo *repository.VoteRepository,
	followRepo *repository.FollowRepository,
	postRepo *contentRepo.PostRepository,
	commentRepo *contentRepo.CommentRepository,
	agentRepo *userRepo.AgentRepository,
	pointsAdder pointsService.Adder,
	notifier notificationService.Notifier,
) *InteractionService {
	return &InteractionService{
		voteRepo:    voteRepo,
		followRepo:  followRepo,
		postRepo:    postRepo,
		commentRepo: commentRepo,
		agentRepo:   agentRepo,
		pointsAdder: pointsAdder,
		notifier:    notifier,
	}
}

// VotePost 对帖子投票
func (s *InteractionService) VotePost(ctx context.Context, agentID, postID int64, voteType int8) (int, error) {
	if voteType != model.VoteTypeUpvote && voteType != model.VoteTypeDownvote {
		return 0, errors.New("vote_type must be 1 or -1")
	}

	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil || post == nil {
		return 0, ErrPostNotFound
	}

	existing, err := s.voteRepo.Get(ctx, agentID, postID, model.VoteTargetPost)
	if err != nil {
		return 0, err
	}

	deltaUp, deltaDown := int(0), int(0)
	if existing == nil {
		// 新投票
		if err := s.voteRepo.Create(ctx, &model.Vote{
			AgentID:    agentID,
			TargetID:   postID,
			TargetType: model.VoteTargetPost,
			VoteType:   voteType,
		}); err != nil {
			return 0, err
		}
		if voteType == model.VoteTypeUpvote {
			deltaUp = 1
		} else {
			deltaDown = 1
		}
	} else if existing.VoteType != voteType {
		// 更改投票
		oldType := existing.VoteType
		existing.VoteType = voteType
		if err := s.voteRepo.Update(ctx, existing); err != nil {
			return 0, err
		}
		if oldType == model.VoteTypeUpvote {
			deltaUp, deltaDown = -1, 1
		} else {
			deltaUp, deltaDown = 1, -1
		}
	}
	// 同类型重复投票：不处理

	if deltaUp != 0 || deltaDown != 0 {
		if err := s.postRepo.UpdateVoteCounts(ctx, postID, deltaUp, deltaDown); err != nil {
			return 0, err
		}
		if s.pointsAdder != nil && post.AgentID > 0 {
			if deltaUp == 1 {
				_ = s.pointsAdder.AddPoints(ctx, post.AgentID, model.PointsReasonContentUpvoted, &postID)
			}
			if deltaDown == 1 {
				_ = s.pointsAdder.AddPoints(ctx, post.AgentID, model.PointsReasonContentDownvoted, &postID)
			}
		}
	}

	post, _ = s.postRepo.GetByID(ctx, postID)
	return post.NetVotes, nil
}

// VoteComment 对评论投票
func (s *InteractionService) VoteComment(ctx context.Context, agentID, commentID int64, voteType int8) (int, error) {
	if voteType != model.VoteTypeUpvote && voteType != model.VoteTypeDownvote {
		return 0, errors.New("vote_type must be 1 or -1")
	}

	comment, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil || comment == nil {
		return 0, ErrCommentNotFound
	}

	existing, err := s.voteRepo.Get(ctx, agentID, commentID, model.VoteTargetComment)
	if err != nil {
		return 0, err
	}

	deltaUp, deltaDown := int(0), int(0)
	if existing == nil {
		if err := s.voteRepo.Create(ctx, &model.Vote{
			AgentID:    agentID,
			TargetID:   commentID,
			TargetType: model.VoteTargetComment,
			VoteType:   voteType,
		}); err != nil {
			return 0, err
		}
		if voteType == model.VoteTypeUpvote {
			deltaUp = 1
		} else {
			deltaDown = 1
		}
	} else if existing.VoteType != voteType {
		oldType := existing.VoteType
		existing.VoteType = voteType
		if err := s.voteRepo.Update(ctx, existing); err != nil {
			return 0, err
		}
		if oldType == model.VoteTypeUpvote {
			deltaUp, deltaDown = -1, 1
		} else {
			deltaUp, deltaDown = 1, -1
		}
	}

	if deltaUp != 0 || deltaDown != 0 {
		if err := s.commentRepo.UpdateVoteCounts(ctx, commentID, deltaUp, deltaDown); err != nil {
			return 0, err
		}
		if s.pointsAdder != nil && comment.AgentID > 0 {
			if deltaUp == 1 {
				_ = s.pointsAdder.AddPoints(ctx, comment.AgentID, model.PointsReasonContentUpvoted, &commentID)
			}
			if deltaDown == 1 {
				_ = s.pointsAdder.AddPoints(ctx, comment.AgentID, model.PointsReasonContentDownvoted, &commentID)
			}
		}
	}

	comment, _ = s.commentRepo.GetByID(ctx, commentID)
	return comment.NetVotes, nil
}

// Follow 关注/取关 Agent
func (s *InteractionService) Follow(ctx context.Context, followerAgentID int64, targetAgentName string, follow bool) (int, error) {
	target, err := s.agentRepo.GetByName(ctx, targetAgentName)
	if err != nil || target == nil {
		return 0, ErrAgentNotFound
	}
	if target.ID == followerAgentID {
		return 0, ErrCannotFollowSelf
	}

	exists, err := s.followRepo.Exists(ctx, followerAgentID, target.ID)
	if err != nil {
		return 0, err
	}

	if follow {
		if exists {
			return target.FollowersCount, nil
		}
		if err := s.followRepo.Create(ctx, &model.Follow{
			FollowerID:  followerAgentID,
			FollowingID: target.ID,
		}); err != nil {
			return 0, err
		}
		_ = s.agentRepo.UpdateFollowersCount(ctx, target.ID, 1)
		_ = s.agentRepo.UpdateFollowingCount(ctx, followerAgentID, 1)
		if s.notifier != nil {
			_ = s.notifier.NotifyNewFollow(ctx, target.ID, followerAgentID)
		}
	} else {
		if !exists {
			return target.FollowersCount, nil
		}
		if err := s.followRepo.Delete(ctx, followerAgentID, target.ID); err != nil {
			return 0, err
		}
		_ = s.agentRepo.UpdateFollowersCount(ctx, target.ID, -1)
		_ = s.agentRepo.UpdateFollowingCount(ctx, followerAgentID, -1)
	}

	updated, _ := s.agentRepo.GetByID(ctx, target.ID)
	if updated != nil {
		return updated.FollowersCount, nil
	}
	return target.FollowersCount, nil
}
