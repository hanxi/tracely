package tracely

// ErrorPayload 错误上报数据结构
type ErrorPayload struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Stack   string `json:"stack"`
	URL     string `json:"url"`
	AppID   string `json:"appId"`
}

// ActivePayload 活跃上报数据结构
type ActivePayload struct {
	AppID    string `json:"appId"`
	UserID   string `json:"userId"`
	Page     string `json:"page"`
	Duration int    `json:"duration"`
}
