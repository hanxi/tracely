package tracely

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// reportTask 上报任务
type reportTask struct {
	url        string
	body       interface{}
	headers    map[string]string
	retryCount int
}

// startQueueWorker 启动异步上报队列消费者
func (c *Client) startQueueWorker() {
	go func() {
		for task := range c.queue {
			c.sendWithRetry(task)
		}
	}()
}

// sendWithRetry 发送请求，失败自动重试
func (c *Client) sendWithRetry(task *reportTask) {
	for i := 0; i < 3; i++ {
		err := c.send(task.url, task.body, task.headers)
		if err == nil {
			return // 成功则返回
		}

		// 失败则等待 1 秒后重试
		time.Sleep(time.Second)
	}
	// 重试 3 次后放弃，不阻塞业务
}

// send 发送 HTTP POST 请求
func (c *Client) send(url string, body interface{}, headers map[string]string) error {
	// 序列化 body 为 JSON
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
