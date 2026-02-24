package repository

import (
	"context"
	"time"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// PostRepository 帖子数据访问层
type PostRepository struct {
	db *gorm.DB
}

// NewPostRepository 创建帖子仓储
func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

// Create 创建帖子
func (r *PostRepository) Create(ctx context.Context, p *model.Post) error {
	return r.db.WithContext(ctx).Create(p).Error
}

// GetByID 根据 ID 查询
func (r *PostRepository) GetByID(ctx context.Context, id int64) (*model.Post, error) {
	var p model.Post
	err := r.db.WithContext(ctx).Preload("Agent").Preload("Community").First(&p, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// List 分页查询帖子，支持多种排序
// sortBy: random, new, top, discussed
// timeRange: hour, day, week, month, year, all（仅 top 时生效）
func (r *PostRepository) List(ctx context.Context, sortBy, timeRange string, limit, offset int) ([]*model.Post, int64, error) {
	// 计算时间范围起点（仅 top 时使用）
	var since time.Time
	if sortBy == "top" && timeRange != "" && timeRange != "all" {
		now := time.Now()
		switch timeRange {
		case "hour":
			since = now.Add(-time.Hour)
		case "day":
			since = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		case "week":
			since = now.AddDate(0, 0, -7)
		case "month":
			since = now.AddDate(0, -1, 0)
		case "year":
			since = now.AddDate(-1, 0, 0)
		}
	}

	// Count 与 Find 使用各自独立的查询链，避免 GORM Statement 复用导致 total 错误
	countQ := r.db.WithContext(ctx).Model(&model.Post{})
	if !since.IsZero() {
		countQ = countQ.Where("created_at >= ?", since)
	}
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	findQ := r.db.WithContext(ctx).Model(&model.Post{}).Preload("Agent").Preload("Community")
	if !since.IsZero() {
		findQ = findQ.Where("created_at >= ?", since)
	}
	switch sortBy {
	case "random":
		if r.db.Dialector != nil && r.db.Dialector.Name() == "mysql" {
			findQ = findQ.Order("RAND()")
		} else {
			findQ = findQ.Order("RANDOM()")
		}
	case "new":
		findQ = findQ.Order("created_at DESC")
	case "top":
		findQ = findQ.Order("net_votes DESC, created_at DESC")
	case "discussed":
		findQ = findQ.Order("comments_count DESC, created_at DESC")
	default:
		findQ = findQ.Order("created_at DESC")
	}

	var posts []*model.Post
	err := findQ.Offset(offset).Limit(limit).Find(&posts).Error
	return posts, total, err
}

// Update 更新帖子
func (r *PostRepository) Update(ctx context.Context, p *model.Post) error {
	return r.db.WithContext(ctx).Save(p).Error
}

// Delete 删除帖子
func (r *PostRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&model.Post{}, id).Error
}

// IncrementCommentsCount 评论数 +1
func (r *PostRepository) IncrementCommentsCount(ctx context.Context, postID int64) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("comments_count", gorm.Expr("comments_count + 1")).Error
}

// DecrementCommentsCount 评论数 -1
func (r *PostRepository) DecrementCommentsCount(ctx context.Context, postID int64) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
		UpdateColumn("comments_count", gorm.Expr("CASE WHEN comments_count > 0 THEN comments_count - 1 ELSE 0 END")).Error
}

// UpdateVoteCounts 更新帖子投票数（供互动服务调用）
func (r *PostRepository) UpdateVoteCounts(ctx context.Context, postID int64, deltaUp, deltaDown int) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", postID).
		Updates(map[string]interface{}{
			"upvotes":   gorm.Expr("upvotes + ?", deltaUp),
			"downvotes": gorm.Expr("downvotes + ?", deltaDown),
			"net_votes": gorm.Expr("net_votes + ? - ?", deltaUp, deltaDown),
		}).Error
}

func (r *PostRepository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.WithContext(ctx).DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
