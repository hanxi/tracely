package tracely

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

// eventActive 内置活跃事件类型常量，与 TS SDK 保持一致
const eventActive = "_active"

// Config 客户端配置
type Config struct {
	AppID     string
	AppSecret string
	Host      string
	Timeout   time.Duration // 默认 5s

	// 心跳上报配置
	EnableHeartbeat   bool              // 是否启用自动心跳上报，默认 false
	HeartbeatInterval time.Duration     // 心跳上报间隔，默认 60s
	InstanceID        string            // 实例标识，为空时自动生成
	Tags              map[string]string // 自定义标签（如 env、version 等）
}

// Client 客户端
type Client struct {
	config     Config
	httpClient *http.Client
	queue      chan *reportTask
	startTime  time.Time // 客户端创建时间，用于计算运行时长
	instanceID string    // 实例唯一标识
}

// New 创建新客户端
func New(config Config) *Client {
	// 设置默认超时
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Second
	}

	// 设置默认心跳间隔
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 60 * time.Second
	}

	// 确定实例 ID
	instanceID := config.InstanceID
	if instanceID == "" {
		instanceID = generateInstanceID()
	}

	client := &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		queue:      make(chan *reportTask, 100), // 缓冲容量 100
		startTime:  time.Now(),
		instanceID: instanceID,
	}

	// 启动异步队列消费者
	client.startQueueWorker()

	// 按需启动心跳上报协程
	if config.EnableHeartbeat {
		client.startHeartbeatWorker()
	}

	return client
}

// generateInstanceID 基于主机名 + 进程 ID + 时间戳生成唯一实例 ID
func generateInstanceID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	pid := os.Getpid()
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%s-%s", hostname, strconv.Itoa(pid), strconv.FormatInt(timestamp, 36))
}

// reportActive 上报活跃数据，参考 TS SDK 的 reportActive 实现
// 使用统一的事件上报接口，事件名为 _active
func (c *Client) reportActive() {
	duration := int64(time.Since(c.startTime).Seconds())

	metadata := map[string]interface{}{
		"instanceId": c.instanceID,
		"duration":   duration,
	}

	// 附加自定义标签
	if len(c.config.Tags) > 0 {
		metadata["tags"] = c.config.Tags
	}

	c.ReportEvent(eventActive, metadata, c.instanceID)
}

// startHeartbeatWorker 启动定时心跳上报协程
func (c *Client) startHeartbeatWorker() {
	go func() {
		// 启动时立即上报一次
		c.reportActive()

		ticker := time.NewTicker(c.config.HeartbeatInterval)
		defer ticker.Stop()

		for range ticker.C {
			c.reportActive()
		}
	}()
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

// ReportEvent 上报事件
func (c *Client) ReportEvent(eventName string, metadata map[string]interface{}, userID string) {
	payload := EventPayload{
		EventName: eventName,
		Metadata:  metadata,
		AppID:     c.config.AppID,
		UserID:    userID,
	}

	// 生成认证头
	headers := buildHeaders(c.config.AppID, c.config.AppSecret)

	// 将任务投入异步队列（队列满时丢弃，不阻塞）
	select {
	case c.queue <- &reportTask{
		url:     c.config.Host + "/report/event",
		body:    payload,
		headers: headers,
	}:
	default:
		// 队列满，直接丢弃
	}
}
