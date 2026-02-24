package repository

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"agent-hub/internal/model"
	"gorm.io/gorm"
)

// SearchRepository 搜索数据访问层（基于 MySQL LIKE + 空格分词，可后续对接 Elasticsearch）
type SearchRepository struct {
	db *gorm.DB
}

// NewSearchRepository 创建搜索仓储
func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

// spaceRe 按空白字符分词
var spaceRe = regexp.MustCompile(`\s+`)

// tokenize 将搜索词按空格拆分为多个 token
// 例："人工 智能 agent" → ["人工", "智能", "agent"]
func tokenize(q string) []string {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	parts := spaceRe.Split(q, -1)
	tokens := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			tokens = append(tokens, p)
		}
	}
	return tokens
}

// applyTokenSearch 将分词结果组装为 WHERE 条件：
//   - 每个 token 必须在至少一个字段中匹配（AND across tokens，OR across fields）
//   - 例：["人工","智能"] 在 [name, bio] 上生成：
//     (name LIKE '%人工%' OR bio LIKE '%人工%') AND (name LIKE '%智能%' OR bio LIKE '%智能%')
func applyTokenSearch(db *gorm.DB, tokens []string, fields []string) *gorm.DB {
	for _, token := range tokens {
		pattern := "%" + token + "%"
		conds := make([]string, 0, len(fields))
		args := make([]interface{}, 0, len(fields))
		for _, field := range fields {
			conds = append(conds, fmt.Sprintf("%s LIKE ?", field))
			args = append(args, pattern)
		}
		db = db.Where(strings.Join(conds, " OR "), args...)
	}
	return db
}

// SearchAgents 按关键词搜索 Agent（name、bio），支持分词
func (r *SearchRepository) SearchAgents(ctx context.Context, q string, limit, offset int) ([]*model.Agent, int64, error) {
	tokens := tokenize(q)
	fields := []string{"name", "bio"}

	countQ := applyTokenSearch(r.db.WithContext(ctx).Model(&model.Agent{}), tokens, fields)
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	findQ := applyTokenSearch(r.db.WithContext(ctx).Model(&model.Agent{}), tokens, fields)
	var list []*model.Agent
	err := findQ.Preload("User").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// SearchPosts 按关键词搜索帖子（title、content），支持分词
func (r *SearchRepository) SearchPosts(ctx context.Context, q string, limit, offset int) ([]*model.Post, int64, error) {
	tokens := tokenize(q)
	fields := []string{"title", "content"}

	countQ := applyTokenSearch(r.db.WithContext(ctx).Model(&model.Post{}), tokens, fields)
	var total int64
	if err := countQ.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	findQ := applyTokenSearch(r.db.WithContext(ctx).Model(&model.Post{}), tokens, fields)
	var list []*model.Post
	err := findQ.Preload("Agent").Preload("Community").Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	return list, total, err
}

// SearchAll 同时搜索 Agent 和帖子，返回各自结果集及总数
// agentLimit/agentOffset 控制 Agent 分页，postLimit/postOffset 控制帖子分页
func (r *SearchRepository) SearchAll(ctx context.Context, q string, agentLimit, agentOffset, postLimit, postOffset int) (
	[]*model.Agent, []*model.Post, int64, int64, error,
) {
	tokens := tokenize(q)
	agentFields := []string{"name", "bio"}
	postFields := []string{"title", "content"}

	// Agent count
	var totalAgents int64
	agentCountQ := applyTokenSearch(r.db.WithContext(ctx).Model(&model.Agent{}), tokens, agentFields)
	if err := agentCountQ.Count(&totalAgents).Error; err != nil {
		return nil, nil, 0, 0, err
	}

	// Agent find
	var agents []*model.Agent
	agentFindQ := applyTokenSearch(r.db.WithContext(ctx).Model(&model.Agent{}), tokens, agentFields)
	if err := agentFindQ.Preload("User").Offset(agentOffset).Limit(agentLimit).Find(&agents).Error; err != nil {
		return nil, nil, 0, 0, err
	}

	// Post count
	var totalPosts int64
	postCountQ := applyTokenSearch(r.db.WithContext(ctx).Model(&model.Post{}), tokens, postFields)
	if err := postCountQ.Count(&totalPosts).Error; err != nil {
		return nil, nil, 0, 0, err
	}

	// Post find
	var posts []*model.Post
	postFindQ := applyTokenSearch(r.db.WithContext(ctx).Model(&model.Post{}), tokens, postFields)
	if err := postFindQ.Preload("Agent").Preload("Community").Order("created_at DESC").Offset(postOffset).Limit(postLimit).Find(&posts).Error; err != nil {
		return nil, nil, 0, 0, err
	}

	return agents, posts, totalAgents, totalPosts, nil
}
