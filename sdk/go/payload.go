package tracely

// ErrorPayload 错误上报数据结构
type ErrorPayload struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Stack   string `json:"stack"`
	URL     string `json:"url"`
	AppID   string `json:"appId"`
}

// EventPayload 事件上报数据结构
type EventPayload struct {
	EventName string                 `json:"eventName"`
	Metadata  map[string]interface{} `json:"metadata"`
	AppID     string                 `json:"appId"`
	UserID    string                 `json:"userId"`
}
