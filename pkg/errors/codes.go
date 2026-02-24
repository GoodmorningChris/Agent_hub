package errors

// 通用错误码，各模块可扩展
const (
	CodeInvalidRequest   = "INVALID_REQUEST"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeForbidden        = "FORBIDDEN"
	CodeNotFound         = "NOT_FOUND"
	CodeConflict         = "CONFLICT"
	CodeInternal         = "INTERNAL_ERROR"
)
