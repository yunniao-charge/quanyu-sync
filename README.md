# 全裕电池数据同步服务

Go 语言实现，定时从全裕 API 拉取数据并接收推送回调，写入 MongoDB。

## 架构

**数据流**:
- **拉取模式**: Cron → Syncer → 全裕 API → MongoDB
- **推送模式**: 全裕 → HTTP Server (/callback/push) → Handler → MongoDB

**核心模块**:
- `internal/sync/syncer.go`: 定时调度器，管理 5 种数据类型的同步任务
- `internal/sync/subscriber.go`: 订阅管理，每 25 分钟续订一次（订阅有效期 30 分钟）
- `internal/storage/`: MongoDB CRUD 操作，所有集合通过唯一索引去重
- `internal/quanyu/client.go`: API 客户端，含签名、重试（3 次，指数退避）
- `internal/device/provider.go`: 设备 UID 列表获取，每 6 小时刷新一次

**增量同步设计**: 每个 UID 独立跟踪同步进度（`sync_states` 集合），单个 UID 失败不影响其他 UID。

## 常用命令

```bash
# 运行服务
go run ./cmd/main.go -config config.yaml

# 构建
go build -o quanyu-battery-sync ./cmd/main.go

# 运行测试
go test ./...
go test -v ./internal/quanyu/...          # 指定包
go test -v -run TestGenerateSign ./internal/quanyu/  # 单个测试

# 运行测试并生成覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out          # 查看覆盖率
go tool cover -html=coverage.out          # HTML 报告

# 依赖管理
go mod download
go mod tidy
```

## MongoDB 集合

数据库: `quanyu_battery`

- 快照类（每 UID 一条，upsert）: `battery_details`, `callback_info`
- 日志类（追加，唯一索引去重）: `battery_history`, `battery_traces`, `battery_events`, `charge_records`, `callback_alarms`, `callback_online`
- `sync_states`: 每 UID 每类型的同步进度

## 配置文件

`config.yaml` 包含所有配置：
- `quanyu`: API 凭证和超时设置
- `mongodb`: 连接配置
- `callback`: HTTP 服务地址
- `sync`: 各数据类型的 cron 表达式和启用状态

## 签名算法

```
signStr = "appid={appid}&nonce_str={nonce_str}&uid={uid}&key={key}"
sign = MD5(signStr).toUpperCase()
```

## 开发流程

详见 [CLAUDE.md](CLAUDE.md)
