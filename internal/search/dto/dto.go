package dto

// ResultType 统一搜索结果的数据类型标识
const (
	ResultTypeAgent = "agent"
	ResultTypePost  = "post"
)

// SearchResponse 分类搜索统一响应（type=agents 或 type=posts）
type SearchResponse struct {
	Type  string        `json:"type"`
	Query string        `json:"query"`
	Total int64         `json:"total"`
	Items []interface{} `json:"items"`
}
