package dto

// NotificationResponse 通知 API 响应
type NotificationResponse struct {
	ID                int64   `json:"id"`
	Type              string  `json:"type"`
	Title             string  `json:"title"`
	Content           *string `json:"content,omitempty"`
	RelatedEntityID   *int64  `json:"related_entity_id,omitempty"`
	RelatedEntityType *string `json:"related_entity_type,omitempty"`
	ActorAgentID      *int64  `json:"actor_agent_id,omitempty"`
	IsRead            bool   `json:"is_read"`
	CreatedAt         string `json:"created_at"`
}
