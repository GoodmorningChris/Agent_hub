package dto

import "agent-hub/internal/model"

// PostResponse 帖子 API 响应
type PostResponse struct {
	ID            int64   `json:"id"`
	AgentID       int64   `json:"agent_id"`
	CommunityID   int64   `json:"community_id"`
	Title         string  `json:"title"`
	Content       *string `json:"content,omitempty"`
	Upvotes       int     `json:"upvotes"`
	Downvotes     int     `json:"downvotes"`
	NetVotes      int     `json:"net_votes"`
	CommentsCount int    `json:"comments_count"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`

	Agent     *AgentBriefResponse `json:"agent,omitempty"`
	Community *CommunityBriefResponse `json:"community,omitempty"`
}

// AgentBriefResponse Agent 简要信息
type AgentBriefResponse struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// CommunityBriefResponse 社区简要信息
type CommunityBriefResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CommentResponse 评论 API 响应
type CommentResponse struct {
	ID        int64   `json:"id"`
	AgentID   int64   `json:"agent_id"`
	PostID    int64   `json:"post_id"`
	Content   string  `json:"content"`
	Upvotes   int     `json:"upvotes"`
	Downvotes int     `json:"downvotes"`
	NetVotes  int     `json:"net_votes"`
	CreatedAt string  `json:"created_at"`

	Agent *AgentBriefResponse `json:"agent,omitempty"`
}

// ToPostResponse 帖子转 API 响应
func ToPostResponse(p *model.Post) PostResponse {
	resp := PostResponse{
		ID:            p.ID,
		AgentID:       p.AgentID,
		CommunityID:   p.CommunityID,
		Title:         p.Title,
		Content:       p.Content,
		Upvotes:       p.Upvotes,
		Downvotes:     p.Downvotes,
		NetVotes:      p.NetVotes,
		CommentsCount: p.CommentsCount,
		CreatedAt:     p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if p.Agent != nil {
		resp.Agent = &AgentBriefResponse{ID: p.Agent.ID, Name: p.Agent.Name, AvatarURL: p.Agent.AvatarURL}
	}
	if p.Community != nil {
		resp.Community = &CommunityBriefResponse{ID: p.Community.ID, Name: p.Community.Name}
	}
	return resp
}

// ToCommentResponse 评论转 API 响应
func ToCommentResponse(c *model.Comment) CommentResponse {
	resp := CommentResponse{
		ID:        c.ID,
		AgentID:   c.AgentID,
		PostID:    c.PostID,
		Content:   c.Content,
		Upvotes:   c.Upvotes,
		Downvotes: c.Downvotes,
		NetVotes:  c.NetVotes,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if c.Agent != nil {
		resp.Agent = &AgentBriefResponse{ID: c.Agent.ID, Name: c.Agent.Name, AvatarURL: c.Agent.AvatarURL}
	}
	return resp
}
