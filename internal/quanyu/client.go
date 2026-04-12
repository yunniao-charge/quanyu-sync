package quanyu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"

	"quanyu-battery-sync/internal/config"

	"go.uber.org/zap"
)

// Client 全裕API客户端
type Client struct {
	config     config.QuanyuConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// 可重试的错误码
var retryableErrnos = map[int]bool{
	500: true,
	502: true,
	503: true,
	504: true,
	408: true,
	429: true,
}

// NewClient 创建新的API客户端
func NewClient(cfg config.QuanyuConfig, logger *zap.Logger) *Client {
	return &Client{
		config: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
		logger: logger,
	}
}

// buildPayload 构建请求体
func (c *Client) buildPayload(uid string, extraParams map[string]any) map[string]any {
	nonceStr := GenerateNonceStr(32)
	timestamp := GetCurrentTimestamp()
	sign := GenerateSign(c.config.AppID, nonceStr, uid, c.config.Key)

	payload := map[string]any{
		"appid":     c.config.AppID,
		"uid":       uid,
		"nonce_str": nonceStr,
		"sign":      sign,
		"time":      timestamp,
	}

	for k, v := range extraParams {
		payload[k] = v
	}

	return payload
}

// makeRequest 发起HTTP请求（含重试）
func (c *Client) makeRequest(ctx context.Context, apiName, endpoint, uid string, extraParams map[string]any) (*QuanyuResponse, error) {
	maxRetries := c.config.MaxRetries
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		payload := c.buildPayload(uid, extraParams)
		jsonData, _ := json.Marshal(payload)

		startTime := time.Now()
		c.logger.Debug("API request",
			zap.String("endpoint", endpoint),
			zap.String("uid", uid),
			zap.Int("attempt", attempt),
		)

		req, err := http.NewRequestWithContext(ctx, "POST", c.config.BaseURL+endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			elapsed := time.Since(startTime)
			lastErr = err
			c.logger.Warn("request error, retrying",
				zap.Int("attempt", attempt),
				zap.Int("max_retries", maxRetries),
				zap.Duration("elapsed", elapsed),
				zap.Error(err),
			)
			if attempt < maxRetries {
				c.backoff(attempt)
				continue
			}
			return nil, fmt.Errorf("请求失败，已重试%d次: %w", maxRetries, err)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		elapsed := time.Since(startTime)

		var result QuanyuResponse
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		// 记录 API 调试日志
		prettyResp, _ := json.MarshalIndent(result, "", "  ")
		logger := c.logger
		if logger != nil {
			logger.Debug("API response",
				zap.String("endpoint", endpoint),
				zap.String("uid", uid),
				zap.Int("attempt", attempt),
				zap.Duration("elapsed", elapsed),
				zap.Int("errno", result.Errno),
			)
		}

		// 检查业务错误码
		if result.Errno != 0 {
			if retryableErrnos[result.Errno] {
				c.logger.Warn("business error, retrying",
					zap.Int("attempt", attempt),
					zap.Int("errno", result.Errno),
					zap.String("errmsg", result.Errmsg),
					zap.Duration("elapsed", elapsed),
				)
				if attempt < maxRetries {
					c.backoff(attempt)
					continue
				}
			}
			c.logger.Error("API business error",
				zap.String("endpoint", endpoint),
				zap.Int("errno", result.Errno),
				zap.String("errmsg", result.Errmsg),
			)
			return &result, nil
		}

		c.logger.Info("API success",
			zap.String("endpoint", endpoint),
			zap.String("uid", uid),
			zap.Int("attempt", attempt),
			zap.Duration("elapsed", elapsed),
		)

		_ = prettyResp // 可用于 LogAPI
		return &result, nil
	}

	return nil, fmt.Errorf("请求失败: %w", lastErr)
}

// backoff 指数退避
func (c *Client) backoff(attempt int) {
	delay := time.Duration(math.Pow(2, float64(attempt-1))) * time.Second
	c.logger.Debug("backoff", zap.Int("attempt", attempt), zap.Duration("delay", delay))
	time.Sleep(delay)
}

// GetBatteryDetail 获取电池详情
func (c *Client) GetBatteryDetail(ctx context.Context, uid string) (*BatteryDetail, error) {
	resp, err := c.makeRequest(ctx, "device-detail", "/sw/api/open/device/detail", uid, nil)
	if err != nil {
		return nil, err
	}
	if resp.Errno != 0 {
		return nil, fmt.Errorf("API错误 [errno=%d]: %s", resp.Errno, resp.Errmsg)
	}

	var detail BatteryDetail
	if err := json.Unmarshal(resp.Data, &detail); err != nil {
		return nil, fmt.Errorf("解析电池详情失败: %w", err)
	}
	return &detail, nil
}

// GetBatteryData 获取电池历史数据
func (c *Client) GetBatteryData(ctx context.Context, uid, startTime, endTime string, last int) (*BatteryDataResponse, error) {
	params := map[string]any{
		"startTime": startTime,
		"endTime":   endTime,
		"last":      last,
	}
	resp, err := c.makeRequest(ctx, "device-data", "/sw/api/open/device/data", uid, params)
	if err != nil {
		return nil, err
	}
	if resp.Errno != 0 {
		return nil, fmt.Errorf("API错误 [errno=%d]: %s", resp.Errno, resp.Errmsg)
	}

	var data BatteryDataResponse
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		return nil, fmt.Errorf("解析电池历史数据失败: %w", err)
	}
	return &data, nil
}

// GetBatteryTrace 获取电池轨迹
func (c *Client) GetBatteryTrace(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*BatteryTraceResponse, error) {
	params := map[string]any{
		"uid":        uid,
		"start_time": startTime,
		"end_time":   endTime,
		"pageNum":    pageNum,
		"pageSize":   pageSize,
	}
	resp, err := c.makeRequest(ctx, "device-trace", "/sw/api/open/battrace", uid, params)
	if err != nil {
		return nil, err
	}
	if resp.Errno != 0 {
		return nil, fmt.Errorf("API错误 [errno=%d]: %s", resp.Errno, resp.Errmsg)
	}

	var trace BatteryTraceResponse
	if err := json.Unmarshal(resp.Data, &trace); err != nil {
		return nil, fmt.Errorf("解析电池轨迹失败: %w", err)
	}
	return &trace, nil
}

// GetBatteryEvents 获取电池事件
func (c *Client) GetBatteryEvents(ctx context.Context, uid, startTime, endTime string, page, limit int) (*BatteryEventResponse, error) {
	params := map[string]any{
		"startTime": startTime,
		"endTime":   endTime,
		"page":      page,
		"limit":     limit,
	}
	resp, err := c.makeRequest(ctx, "device-event", "/sw/api/open/device/eventData", uid, params)
	if err != nil {
		return nil, err
	}
	if resp.Errno != 0 {
		return nil, fmt.Errorf("API错误 [errno=%d]: %s", resp.Errno, resp.Errmsg)
	}

	var events BatteryEventResponse
	if err := json.Unmarshal(resp.Data, &events); err != nil {
		return nil, fmt.Errorf("解析电池事件失败: %w", err)
	}
	return &events, nil
}

// GetChargeRecords 获取充放电记录
func (c *Client) GetChargeRecords(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*ChargeDataResponse, error) {
	params := map[string]any{
		"beginStart": beginStart,
		"beginEnd":   beginEnd,
		"page":       page,
		"limit":      limit,
	}
	resp, err := c.makeRequest(ctx, "device-charge", "/sw/api/open/device/chargeData", uid, params)
	if err != nil {
		return nil, err
	}
	if resp.Errno != 0 {
		return nil, fmt.Errorf("API错误 [errno=%d]: %s", resp.Errno, resp.Errmsg)
	}

	var records ChargeDataResponse
	if err := json.Unmarshal(resp.Data, &records); err != nil {
		return nil, fmt.Errorf("解析充放电记录失败: %w", err)
	}
	return &records, nil
}

// SubscribeV2 订阅设备数据推送
func (c *Client) SubscribeV2(ctx context.Context, uid string, list []string, subData []string, notifyURL string) (*QuanyuResponse, error) {
	params := map[string]any{
		"list":       list,
		"sub_data":   subData,
		"notify_url": notifyURL,
	}
	resp, err := c.makeRequest(ctx, "subscribe", "/sw/api/open/subscribeV2", uid, params)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
