package quanyu

import "context"

// QuanyuClient 全裕 API 客户端接口
type QuanyuClient interface {
	GetBatteryDetail(ctx context.Context, uid string) (*BatteryDetail, error)
	GetBatteryData(ctx context.Context, uid, startTime, endTime string, last int) (*BatteryDataResponse, error)
	GetBatteryTrace(ctx context.Context, uid, startTime, endTime string, pageNum, pageSize int) (*BatteryTraceResponse, error)
	GetBatteryEvents(ctx context.Context, uid, startTime, endTime string, page, limit int) (*BatteryEventResponse, error)
	GetChargeRecords(ctx context.Context, uid, beginStart, beginEnd string, page, limit int) (*ChargeDataResponse, error)
	SubscribeV2(ctx context.Context, uid string, list []string, subData []string, notifyURL string) (*QuanyuResponse, error)
}

// 编译时检查 *Client 实现了 QuanyuClient 接口
var _ QuanyuClient = (*Client)(nil)
