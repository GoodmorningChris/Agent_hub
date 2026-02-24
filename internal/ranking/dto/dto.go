package dto

import (
	"agent-hub/internal/model"
	userDto "agent-hub/internal/user/dto"
	contentDto "agent-hub/internal/content/dto"
)

// LeaderboardResponse 排行榜统一响应（根据 type 不同，items 为 agents 或 posts）
type LeaderboardResponse struct {
	Type  string        `json:"type"`
	Items []interface{} `json:"items"`
}

// AgentRankItem 积分榜/影响力榜单项
type AgentRankItem struct {
	Rank  int                           `json:"rank"`
	Agent userDto.AgentPublicResponse   `json:"agent"`
}

// PostRankItem 内容榜单项
type PostRankItem struct {
	Rank int                     `json:"rank"`
	Post contentDto.PostResponse `json:"post"`
}

// ToAgentRankItems 带排名的 Agent 列表
func ToAgentRankItems(agents []*model.Agent) []AgentRankItem {
	items := make([]AgentRankItem, len(agents))
	for i, a := range agents {
		items[i] = AgentRankItem{
			Rank:  i + 1,
			Agent: userDto.ToAgentPublicResponse(a),
		}
	}
	return items
}

// ToPostRankItems 带排名的帖子列表
func ToPostRankItems(posts []*model.Post) []PostRankItem {
	items := make([]PostRankItem, len(posts))
	for i, p := range posts {
		items[i] = PostRankItem{
			Rank: i + 1,
			Post: contentDto.ToPostResponse(p),
		}
	}
	return items
}
