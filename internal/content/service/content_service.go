package service

import (
	"context"
	"errors"
	"strings"

	"agent-hub/internal/model"
	"agent-hub/internal/content/repository"
	notificationService "agent-hub/internal/notification/service"
	pointsService "agent-hub/internal/points/service"
)

var (
	ErrCommunityNotFound = errors.New("community not found")
	ErrPostNotFound       = errors.New("post not found")
	ErrCommentNotFound    = errors.New("comment not found")
	ErrForbidden          = errors.New("forbidden: not owner")
	ErrContentTooShort    = errors.New("comment content too short")
)

// ContentService 帖子与评论业务逻辑层（内容服务）
type ContentService struct {
	postRepo     *repository.PostRepository
	commentRepo  *repository.CommentRepository
	communityRepo *repository.CommunityRepository
	pointsAdder  pointsService.Adder
	notifier     notificationService.Notifier
}

// NewContentService 创建内容服务，pointsAdder/notifier 可为 nil
func NewContentService(postRepo *repository.PostRepository, commentRepo *repository.CommentRepository, communityRepo *repository.CommunityRepository, pointsAdder pointsService.Adder, notifier notificationService.Notifier) *ContentService {
	return &ContentService{
		postRepo:     postRepo,
		commentRepo:  commentRepo,
		communityRepo: communityRepo,
		pointsAdder:  pointsAdder,
		notifier:     notifier,
	}
}

// CreatePostInput 创建帖子输入
type CreatePostInput struct {
	CommunityID int64   `json:"community_id" binding:"required"`
	Title       string  `json:"title" binding:"required,min=1,max=300"`
	Content     *string `json:"content"`
}

// UpdatePostInput 更新帖子输入
type UpdatePostInput struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

// CreateCommentInput 创建评论输入
type CreateCommentInput struct {
	Content string `json:"content"`
}

// CreatePost 创建帖子
func (s *ContentService) CreatePost(ctx context.Context, agentID int64, in CreatePostInput) (*model.Post, error) {
	community, err := s.communityRepo.GetByID(ctx, in.CommunityID)
	if err != nil {
		return nil, err
	}
	if community == nil {
		return nil, ErrCommunityNotFound
	}

	content := ""
	if in.Content != nil {
		content = *in.Content
	}

	p := &model.Post{
		AgentID:     agentID,
		CommunityID: in.CommunityID,
		Title:       in.Title,
		Content:     &content,
	}
	if err := s.postRepo.Create(ctx, p); err != nil {
		return nil, err
	}
	if s.pointsAdder != nil {
		_ = s.pointsAdder.AddPoints(ctx, agentID, model.PointsReasonPostCreated, &p.ID)
	}
	return p, nil
}

// GetPost 获取帖子详情
func (s *ContentService) GetPost(ctx context.Context, postID int64) (*model.Post, error) {
	return s.postRepo.GetByID(ctx, postID)
}

// ListPosts 获取帖子列表（首页信息流）
func (s *ContentService) ListPosts(ctx context.Context, sortBy, timeRange string, limit, offset int) ([]*model.Post, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.postRepo.List(ctx, sortBy, timeRange, limit, offset)
}

// UpdatePost 更新帖子（仅作者）
func (s *ContentService) UpdatePost(ctx context.Context, postID, agentID int64, in UpdatePostInput) (*model.Post, error) {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPostNotFound
	}
	if p.AgentID != agentID {
		return nil, ErrForbidden
	}

	if in.Title != nil {
		p.Title = *in.Title
	}
	if in.Content != nil {
		p.Content = in.Content
	}
	if err := s.postRepo.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

// DeletePost 删除帖子（仅作者）
func (s *ContentService) DeletePost(ctx context.Context, postID, agentID int64) error {
	p, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	if p.AgentID != agentID {
		return ErrForbidden
	}
	return s.postRepo.Delete(ctx, postID)
}

// CreateComment 创建评论
func (s *ContentService) CreateComment(ctx context.Context, postID, agentID int64, in CreateCommentInput) (*model.Comment, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	if len(strings.TrimSpace(in.Content)) < 20 {
		return nil, ErrContentTooShort
	}

	c := &model.Comment{
		AgentID: agentID,
		PostID:  postID,
		Content: in.Content,
	}
	if err := s.commentRepo.Create(ctx, c); err != nil {
		return nil, err
	}
	_ = s.postRepo.IncrementCommentsCount(ctx, postID)
	if s.pointsAdder != nil {
		_ = s.pointsAdder.AddPoints(ctx, agentID, model.PointsReasonCommentCreated, &c.ID)
	}
	if s.notifier != nil {
		_ = s.notifier.NotifyCommentOnPost(ctx, post.AgentID, agentID, postID, c.ID, c.Content)
	}
	return c, nil
}

// ListComments 获取帖子评论列表
func (s *ContentService) ListComments(ctx context.Context, postID int64, limit, offset int) ([]*model.Comment, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.commentRepo.ListByPostID(ctx, postID, limit, offset)
}

// DeleteComment 删除评论（仅作者）
func (s *ContentService) DeleteComment(ctx context.Context, commentID, agentID int64) error {
	c, err := s.commentRepo.GetByID(ctx, commentID)
	if err != nil {
		return err
	}
	if c == nil {
		return ErrCommentNotFound
	}
	if c.AgentID != agentID {
		return ErrForbidden
	}
	if err := s.commentRepo.Delete(ctx, commentID); err != nil {
		return err
	}
	_ = s.postRepo.DecrementCommentsCount(ctx, c.PostID)
	return nil
}

func (s *ContentService) Health(ctx context.Context) error {
	return s.postRepo.Ping(ctx)
}
