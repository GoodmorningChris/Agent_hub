package repository

import (
	"context"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// CommentRepository 评论数据访问层
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 创建评论仓储
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create 创建评论
func (r *CommentRepository) Create(ctx context.Context, c *model.Comment) error {
	return r.db.WithContext(ctx).Create(c).Error
}

// GetByID 根据 ID 查询
func (r *CommentRepository) GetByID(ctx context.Context, id int64) (*model.Comment, error) {
	var c model.Comment
	err := r.db.WithContext(ctx).Preload("Agent").First(&c, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &c, nil
}

// ListByPostID 按帖子 ID 分页查询评论，按净票数排序
func (r *CommentRepository) ListByPostID(ctx context.Context, postID int64, limit, offset int) ([]*model.Comment, int64, error) {
	// Count 与 Find 使用各自独立的查询链，避免 GORM Statement 复用导致 total 错误
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ?", postID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var comments []*model.Comment
	err := r.db.WithContext(ctx).Model(&model.Comment{}).Where("post_id = ?", postID).
		Preload("Agent").Order("net_votes DESC, created_at ASC").
		Offset(offset).Limit(limit).Find(&comments).Error
	return comments, total, err
}

// Delete 删除评论
func (r *CommentRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Comment{}, id).Error
}

// UpdateVoteCounts 更新评论投票数（供互动服务调用）
func (r *CommentRepository) UpdateVoteCounts(ctx context.Context, commentID int64, deltaUp, deltaDown int) error {
	return r.db.WithContext(ctx).Model(&model.Comment{}).Where("id = ?", commentID).
		Updates(map[string]interface{}{
			"upvotes":   gorm.Expr("upvotes + ?", deltaUp),
			"downvotes": gorm.Expr("downvotes + ?", deltaDown),
			"net_votes": gorm.Expr("net_votes + ? - ?", deltaUp, deltaDown),
		}).Error
}

func (r *CommentRepository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
