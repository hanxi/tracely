package tracely

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// reportTask 上报任务
type reportTask struct {
	url  string
	body interface{}
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
		err := c.send(task.url, task.body)
		if err == nil {
			return
		}
		slog.Error("failed to send request", "url", task.url, "err", err)
		time.Sleep(time.Second)
	}
}

// send 发送 HTTP POST 请求（每次调用重新生成签名头，避免重试时 nonce 重放）
func (c *Client) send(url string, body interface{}) error {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range buildHeaders(c.config.AppID, c.config.AppSecret) {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
