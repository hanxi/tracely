package tracely

import (
	"net/http"
	"time"
)

// Config 客户端配置
type Config struct {
	AppID     string
	AppSecret string
	Host      string
	Timeout   time.Duration // 默认 5s
}

// Client 客户端
type Client struct {
	config     Config
	httpClient *http.Client
	queue      chan *reportTask
}

// New 创建新客户端
func New(config Config) *Client {
	// 设置默认超时
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		queue: make(chan *reportTask, 100), // 缓冲容量 100
	}

	// 启动异步队列消费者
	client.startQueueWorker()

	return client
}

// ReportError 上报错误
func (c *Client) ReportError(payload ErrorPayload) {
	// 自动填充 AppID
	payload.AppID = c.config.AppID

	// 生成认证头
	headers := buildHeaders(c.config.AppID, c.config.AppSecret)

	// 将任务投入异步队列（队列满时丢弃，不阻塞）
	select {
	case c.queue <- &reportTask{
		url:     c.config.Host + "/report/error",
		body:    payload,
		headers: headers,
	}:
	default:
		// 队列满，直接丢弃
	}
}

// ReportActive 上报活跃
func (c *Client) ReportActive(payload ActivePayload) {
	// 自动填充 AppID
	payload.AppID = c.config.AppID

	// 生成认证头
	headers := buildHeaders(c.config.AppID, c.config.AppSecret)

	// 将任务投入异步队列（队列满时丢弃，不阻塞）
	select {
	case c.queue <- &reportTask{
		url:     c.config.Host + "/report/active",
		body:    payload,
		headers: headers,
	}:
	default:
		// 队列满，直接丢弃
	}
}
