# 全裕电池数据同步服务

Go 语言实现，定时从全裕 API 拉取电池数据并接收推送回调，写入 MongoDB。

## 数据流

- **拉取模式**: Cron 定时 → Syncer → 全裕 API → MongoDB
- **推送模式**: 全裕推送 → HTTP Server `/callback/push` → MongoDB

## 快速开始

```bash
# 1. 复制配置文件并填写
cp config.yaml.example config.yaml
# 编辑 config.yaml，填入 API 凭证、MongoDB 地址、回调公网地址

# 2. 运行
go run ./cmd/main.go -config config.yaml

# 3. 编译
go build -o quanyu-battery-sync ./cmd/
```

## 服务器部署

```bash
# 首次部署
git clone <repo-url> /opt/quanyu-sync
cd /opt/quanyu-sync
cp config.yaml.example config.yaml
# 编辑 config.yaml
cp scripts/quanyu-sync.service /etc/systemd/system/
systemctl daemon-reload && systemctl enable quanyu-sync

# 后续更新
bash scripts/deploy.sh    # git pull + build + restart
bash scripts/restart.sh   # 重启（修改配置后）
bash scripts/stop.sh      # 停止
```

## 测试

```bash
go test ./...                                        # 单元测试
go test -tags=integration -v ./tests/integration/... # 集成测试（需 MongoDB + API）
```

## MongoDB 集合

数据库 `quanyu_battery`，集合和索引在启动时自动创建。

| 集合 | 说明 |
|------|------|
| battery_details | 设备快照（每设备一条，pull + callback 更新） |
| battery_history | 历史时序数据 |
| battery_traces | GPS 轨迹 |
| battery_events | 告警事件 |
| charge_records | 充放电记录 |
| callback_alarms | 告警回调 |
| callback_online | 在线状态回调 |
| sync_states | 同步进度跟踪 |
