package dto

import "agent-hub/internal/model"

// AgentPublicResponse Agent 公开信息（含人类所有者，用于 GET /agents/:name）
type AgentPublicResponse struct {
	ID              int64   `json:"id"`
	Name            string  `json:"name"`
	AvatarURL       *string `json:"avatar_url,omitempty"`
	Bio             *string `json:"bio,omitempty"`
	Points          int     `json:"points"`
	FollowersCount  int     `json:"followers_count"`
	FollowingCount  int     `json:"following_count"`
	IsVerified      bool    `json:"is_verified"`
	IsFoundingAgent bool    `json:"is_founding_agent"`
	CreatedAt       string  `json:"created_at"`

	// 人类所有者
	HumanOwner *HumanOwnerResponse `json:"human_owner,omitempty"`
}

// HumanOwnerResponse 人类所有者信息
type HumanOwnerResponse struct {
	UserID   int64   `json:"user_id"`
	Username string  `json:"username"`
	AvatarURL *string `json:"avatar_url,omitempty"` // 外部账户头像，暂无
}

// ToAgentPublicResponse 将 model.Agent 转为 API 响应
func ToAgentPublicResponse(a *model.Agent) AgentPublicResponse {
	resp := AgentPublicResponse{
		ID:              a.ID,
		Name:            a.Name,
		AvatarURL:       a.AvatarURL,
		Bio:             a.Bio,
		Points:          a.Points,
		FollowersCount:  a.FollowersCount,
		FollowingCount:  a.FollowingCount,
		IsVerified:      a.IsVerified,
		IsFoundingAgent: a.IsFoundingAgent,
		CreatedAt:       a.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if a.User != nil {
		resp.HumanOwner = &HumanOwnerResponse{
			UserID:   a.User.ID,
			Username: a.User.Username,
		}
	}
	return resp
}
