# quanyu-battery-sync 全裕电池数据同步服务 - 设计方案

## 1. 项目概述

全裕（盛维智联）电池 IoT 数据同步服务。Go 语言实现，独立后台运行，定时从全裕 API 拉取电池数据并接收推送回调，所有数据写入 MongoDB 供其他服务读取。

### 1.1 核心能力

- **定时拉取**：按 cron 配置周期性调用全裕 API 获取电池数据
- **订阅回调**：分批订阅设备，内置 HTTP 服务接收全裕的实时推送（info/alarm/online）
- **增量同步**：每个 uid 独立跟踪同步时间点，配合 MongoDB 唯一索引去重
- **自动翻页**：分页接口自动遍历所有页面，不遗漏数据

### 1.2 排除的接口

以下接口**不使用**：
- `/sw/api/open/batinfo` — 单个电池查询（用 device/detail 替代）
- `/sw/api/open/batinfosV2` — 批量电池查询
- `/sw/api/open/lowElectricity` — 低电设备查询
- `/sw/api/open/deviceList` — 设备列表查询
- `/sw/api/open/batcharge` — 充电控制
- `/sw/api/open/batdischarge` — 放电控制

### 1.3 使用的接口

| 端点 | 用途 | 同步方式 | 数据类型 |
|------|------|----------|----------|
| `/sw/api/open/device/detail` | 电池详情 | 定时拉取 | 快照 |
| `/sw/api/open/device/data` | 历史时序数据 | 定时拉取 | 日志 |
| `/sw/api/open/battrace` | 电池轨迹 | 定时拉取 | 日志 |
| `/sw/api/open/device/eventData` | 电池事件 | 定时拉取 | 日志 |
| `/sw/api/open/device/chargeData` | 充放电记录 | 定时拉取 | 日志 |
| `/sw/api/open/subscribeV2` | 数据订阅 | 定时续订（30min） | — |
| 回调: info | 电池状态推送 | 被动接收 | 快照 |
| 回调: alarm | 告警推送 | 被动接收 | 日志 |
| 回调: online | 上下线事件 | 被动接收 | 日志 |

---

## 2. 技术栈

| 组件 | 选择 | 说明 |
|------|------|------|
| 语言 | Go 1.22+ | — |
| 定时任务 | robfig/cron/v3 | 与现有项目一致 |
| HTTP 客户端 | net/http + 标准库 | 全裕 API 调用 |
| HTTP 服务 | net/http | 回调接收 |
| 数据库 | MongoDB (go.mongodb.org/mongo-driver/v2) | 数据存储 |
| 日志 | zap + lumberjack | 结构化日志 + 文件轮转 |
| 配置 | gopkg.in/yaml.v3 | YAML 配置文件 |
| 测试 | testing + testify | 标准库 + 断言库 |

---

## 3. 项目结构

```
quanyu-battery-sync/
├── cmd/
│   └── main.go                          # 入口：加载配置、初始化依赖、启动服务
├── internal/
│   ├── config/
│   │   └── config.go                    # 配置结构定义与加载
│   ├── logger/
│   │   └── logger.go                    # zap 日志初始化
│   ├── quanyu/
│   │   ├── client.go                    # 全裕 API HTTP 客户端（含重试）
│   │   ├── models.go                    # API 请求/响应数据模型
│   │   ├── sign.go                      # 签名算法（MD5）
│   │   └── sign_test.go                 # 签名单测
│   ├── storage/
│   │   ├── mongo.go                     # MongoDB 连接管理
│   │   ├── battery_detail.go            # 电池详情（快照）CRUD
│   │   ├── battery_history.go           # 历史时序数据 CRUD
│   │   ├── battery_trace.go             # 轨迹数据 CRUD
│   │   ├── battery_event.go             # 事件数据 CRUD
│   │   ├── charge_record.go             # 充放电记录 CRUD
│   │   ├── callback_info.go             # info 回调数据存储
│   │   ├── callback_alarm.go            # alarm 回调数据存储
│   │   ├── callback_online.go           # online 回调数据存储
│   │   └── sync_state.go                # 同步状态管理（每 uid 时间点）
│   ├── sync/
│   │   ├── syncer.go                    # 同步调度器（定时拉取）
│   │   ├── detail.go                    # detail 同步逻辑
│   │   ├── history.go                   # 历史数据同步逻辑（含分页）
│   │   ├── trace.go                     # 轨迹同步逻辑（含分页）
│   │   ├── event.go                     # 事件同步逻辑（含分页）
│   │   ├── charge.go                    # 充放电同步逻辑（含分页）
│   │   └── subscriber.go               # 订阅管理（分批订阅 + 续订）
│   ├── callback/
│   │   ├── handler.go                   # HTTP 回调处理
│   │   └── models.go                    # 回调数据模型
│   └── device/
│       └── provider.go                  # 设备 UID 列表获取（外部 API）
├── tests/
│   ├── integration/
│   │   ├── client_test.go               # API 客户端集成测试（mock server）
│   │   └── storage_test.go              # MongoDB 集成测试
│   └── testdata/
│       └── *.json                       # 测试用 mock 数据
├── config.yaml                          # 配置文件
├── go.mod
├── go.sum
├── start.bat
└── start.sh
```

---

## 4. 配置设计

### 4.1 config.yaml 结构

```yaml
# 全裕 API 配置
quanyu:
  appid: "ZTvK1rE8jQ9wJ3vJ4oT4qF3eS1cJ8vX5"
  key: "wT0dD7uQ3cB3pV2fZ7dA3sD7eZ7uP4uE"
  base_url: "https://iot.xtrunc.com"
  timeout: 30s
  max_retries: 3

# 设备列表获取
device_api:
  url: "https://api.yunniaotech.com/api/admin/devices/uids"
  refresh_interval: 6h

# MongoDB 配置
mongodb:
  uri: "mongodb://localhost:27017"
  database: "quanyu_battery"
  username: ""
  password: ""

# 回调服务配置
callback:
  listen_addr: ":8080"
  notifyurl_base: "https://myserver.example.com:8080"

# 订阅配置
subscribe:
  batch_size: 20
  renew_interval: 25m   # 25 分钟续订（留 5 分钟余量）
  sub_data:
    - "info"
    - "alarm"
    - "online"

# 日志配置
log:
  dir: "logs"
  level: "info"          # debug/info/warn/error
  max_size: 100          # MB
  max_backups: 30
  max_age: 30            # days
  api_debug_log: true    # 是否记录 API 请求/响应详细日志

# 同步任务配置
sync:
  detail:
    cron: "0 */2 * * * *"    # 每 2 分钟
    enabled: true
    time_range: ""            # 不需要时间范围，获取最新状态
  history_data:
    cron: "0 */5 * * * *"    # 每 5 分钟
    enabled: true
    default_range: 1h         # 默认同步最近 1 小时
  trace:
    cron: "0 */5 * * * *"    # 每 5 分钟
    enabled: true
    default_range: 1h
  event:
    cron: "0 */5 * * * *"    # 每 5 分钟
    enabled: true
    default_range: 1h
  charge_data:
    cron: "0 */10 * * * *"   # 每 10 分钟
    enabled: true
    default_range: 2h
```

### 4.2 配置结构体

```go
type Config struct {
    Quanyu     QuanyuConfig     `yaml:"quanyu"`
    DeviceAPI  DeviceAPIConfig  `yaml:"device_api"`
    MongoDB    MongoDBConfig    `yaml:"mongodb"`
    Callback   CallbackConfig   `yaml:"callback"`
    Subscribe  SubscribeConfig  `yaml:"subscribe"`
    Log        LogConfig        `yaml:"log"`
    Sync       SyncConfig       `yaml:"sync"`
}

type QuanyuConfig struct {
    AppID      string        `yaml:"appid"`
    Key        string        `yaml:"key"`
    BaseURL    string        `yaml:"base_url"`
    Timeout    time.Duration `yaml:"timeout"`
    MaxRetries int           `yaml:"max_retries"`
}

type MongoDBConfig struct {
    URI      string `yaml:"uri"`
    Database string `yaml:"database"`
    Username string `yaml:"username"`
    Password string `yaml:"password"`
}

type CallbackConfig struct {
    ListenAddr    string `yaml:"listen_addr"`
    NotifyURLBase string `yaml:"notifyurl_base"`
}

type SubscribeConfig struct {
    BatchSize     int           `yaml:"batch_size"`
    RenewInterval time.Duration `yaml:"renew_interval"`
    SubData       []string      `yaml:"sub_data"`
}

type SyncTaskConfig struct {
    Cron         string        `yaml:"cron"`
    Enabled      bool          `yaml:"enabled"`
    TimeRange    time.Duration `yaml:"time_range"`     // detail 不需要
    DefaultRange time.Duration `yaml:"default_range"`  // 历史数据默认时间范围
}

type SyncConfig struct {
    Detail     SyncTaskConfig `yaml:"detail"`
    HistoryData SyncTaskConfig `yaml:"history_data"`
    Trace      SyncTaskConfig `yaml:"trace"`
    Event      SyncTaskConfig `yaml:"event"`
    ChargeData SyncTaskConfig `yaml:"charge_data"`
}
```

---

## 5. MongoDB 集合设计

### 5.1 数据库：`quanyu_battery`

### 5.2 集合与索引

#### `battery_details` — 电池详情（快照，每 uid 一条）

```javascript
{
    uid: "8600123456",          // 电池 UID（唯一键）
    sn: "SN001",                // 序列号
    soc: 85,                    // 电量百分比
    voltage: "52.3",            // 总电压 V
    online_status: 1,           // 在线状态
    devstate: 2,                // 充电状态: 0-未知 1-空闲 2-充电中
    cell_voltage: [3265, 3270], // 电芯电压数组 mV
    device_bms1: [25.3],        // 温度数组1
    device_bms2: [24.8],        // 温度数组2
    last_pos: "120.123,30.456", // 位置
    remain: "85%",              // 剩余电量
    current: 5.2,               // 电流 0.1A
    charge: 0,                  // 充电MOS
    discharge: 0,               // 放电MOS
    bat_time: "2024-01-01 12:00:00",
    temp_mos: 30,               // MOS温度
    temp_env: 25,               // 环境温度
    temp_c1: 26,                // 电芯温度1
    temp_c2: 27,                // 电芯温度2
    cap: "20Ah",                // 电池容量
    imsi: "460001234567890",
    imei: "861234567890123",
    signal: "18",
    devtype: 0,                 // 设备类型
    loc: "120.123,30.456",
    loc_time: "2024-01-01 12:00:00",
    n: "30.456",                // 纬度
    e: "120.123",               // 经度
    // ... 其他字段（extra: allow）
    updated_at: ISODate("..."), // 最后更新时间
    sync_source: "pull"         // pull | callback
}
```

**索引**：
- `{ uid: 1 }` unique — 覆盖更新依据
- `{ updated_at: -1 }` — 按更新时间查询

#### `battery_history` — 历史时序数据（日志）

```javascript
{
    uid: "8600123456",
    timestamp: ISODate("..."),          // 时序时间点
    time_scale: [1704067200, ...],      // 原始时间戳数组
    device_bms1: [25.3, ...],           // 温度数组
    device_bms2: [24.8, ...],
    device_ct1: [],                     // 电流数组1
    device_ct2: [],                     // 电流数组2
    device_av: [],                      // 电压数组
    device_cg: [],                      // 功率数组
    device_cc: [],                      // 电量数组
    device_sa: [],                      // 放电功率数组
    device_rem: [],                     // 剩余电量数组
    device_core_v: [],                  // 电芯电压数组
    vvv: "52.3",                        // 电压字符串
    cap: "20Ah",
    rem: "15.6",
    signal: "18",
    synced_at: ISODate("...")           // 入库时间
}
```

**索引**：
- `{ uid: 1, timestamp: 1 }` unique — 去重依据
- `{ uid: 1, synced_at: -1 }` — 按同步时间查询

#### `battery_traces` — 轨迹数据（日志）

```javascript
{
    uid: "8600123456",
    loc: "120.123,30.456",
    loc_time: "20240101120000",         // 原始时间格式
    synced_at: ISODate("...")
}
```

**索引**：
- `{ uid: 1, loc_time: 1 }` unique — 去重依据
- `{ uid: 1, synced_at: -1 }`

#### `battery_events` — 事件数据（日志）

```javascript
{
    uid: "8600123456",
    alarm: "bms_high_temp",             // 告警代码
    type: 1,                            // 0-恢复 1-产生
    time: "2024-01-01 12:00:00",        // 事件时间
    v1: "...", v2: "...",               // 扩展字段
    synced_at: ISODate("...")
}
```

**索引**：
- `{ uid: 1, alarm: 1, time: 1 }` unique — 去重依据
- `{ uid: 1, synced_at: -1 }`

#### `charge_records` — 充放电记录（日志）

```javascript
{
    uid: "8600123456",
    device_id: "...",
    charge_begin: "2024-01-01 10:00:00",
    charge_end: "2024-01-01 12:00:00",
    begin_soc: 20,
    end_soc: 85,
    charge_dwh: 1.5,
    charge_dah: 30,
    acc_ah: 1500,
    drive_miles: 50,
    idx_auto: 12345,                    // 全裕自增 ID
    // ... 其他字段
    synced_at: ISODate("...")
}
```

**索引**：
- `{ uid: 1, idx_auto: 1 }` unique — 去重依据（idx_auto 是全裕平台的唯一自增 ID）
- `{ uid: 1, charge_begin: -1 }` — 按充电开始时间查询

#### `callback_info` — info 推送数据（快照）

```javascript
{
    uid: "8600123456",
    devtype: 0,
    sn: "SN001",
    loc: "120.123,30.456",
    remain: 85,                         // 注意：回调中 remain 是 int
    online: 1,
    voltage: 52300,                     // mV
    charge: 0,
    discharge: 0,
    bat_time: "2024-01-01 12:00:00",
    received_at: ISODate("..."),
    appid: "ZTvK..."
}
```

**索引**：
- `{ uid: 1 }` unique — 覆盖更新（快照）
- `{ received_at: -1 }`

> **注意**：收到 info 回调后，同时更新 `battery_details` 中对应的字段。

#### `callback_alarms` — 告警推送（日志）

```javascript
{
    uid: "8600123456",
    alarm: "bms_high_temp",
    type: 1,                            // 0-恢复 1-产生
    time: "2024-01-01 12:00:00",
    info_obj: { ... },                  // 额外信息
    received_at: ISODate("..."),
    appid: "ZTvK..."
}
```

**索引**：
- `{ uid: 1, alarm: 1, time: 1 }` unique — 去重依据
- `{ received_at: -1 }`

#### `callback_online` — 上下线事件（日志）

```javascript
{
    uid: "8600123456",
    online: 1,                          // 1-上线 0-下线
    time: "2024-01-01 12:00:00",
    received_at: ISODate("..."),
    appid: "ZTvK..."
}
```

**索引**：
- `{ uid: 1, time: 1 }` unique — 去重依据
- `{ received_at: -1 }`

#### `sync_states` — 同步状态（每 uid 每类型一条）

```javascript
{
    uid: "8600123456",
    sync_type: "detail",                // detail/history_data/trace/event/charge_data
    last_sync_time: "2024-01-01 12:00:00",  // 最后成功同步的时间点（作为下次查询的起始时间）
    last_success_at: ISODate("..."),    // 最后成功同步的系统时间
    sync_count: 150,                    // 累计同步次数
    error_count: 2,                     // 累计错误次数
    last_error: "timeout",              // 最后一次错误信息
    last_error_at: ISODate("...")       // 最后一次错误时间
}
```

**索引**：
- `{ uid: 1, sync_type: 1 }` unique — 每类型每 uid 一条
- `{ sync_type: 1, last_sync_time: 1 }` — 按类型查询同步进度

---

## 6. 核心模块设计

### 6.1 全裕 API 客户端（`internal/quanyu/client.go`）

复用现有 quanyu-sync 的核心设计，修正细节：

```go
type Client struct {
    config     QuanyuConfig
    httpClient *http.Client
    logger     *zap.Logger
}

// 核心方法
func (c *Client) MakeRequest(ctx context.Context, endpoint, uid string, extraParams map[string]any) (*QuanyuResponse, error)
func (c *Client) GetBatteryDetail(ctx context.Context, uid string) (*BatteryDetail, error)
func (c *Client) GetBatteryData(ctx context.Context, uid string, startTime, endTime string, last int) (*BatteryDataResponse, error)
func (c *Client) GetBatteryTrace(ctx context.Context, uid string, params TraceParams) (*BatteryTraceResponse, error)
func (c *Client) GetBatteryEvents(ctx context.Context, uid string, params EventParams) (*BatteryEventResponse, error)
func (c *Client) GetChargeRecords(ctx context.Context, uid string, params ChargeParams) (*ChargeDataResponse, error)
func (c *Client) SubscribeV2(ctx context.Context, uid string, list []string, subData []string, notifyURL string) (*QuanyuResponse, error)
```

**重试机制**：
- 最大重试 3 次
- 可重试错误码：500, 502, 503, 504, 408, 429
- 每次重试重新生成签名（nonce_str + timestamp）
- 指数退避：1s, 2s, 4s
- 超时 30s

### 6.2 签名算法（`internal/quanyu/sign.go`）

与现有代码完全一致：

```
signStr = "appid={appid}&nonce_str={nonce_str}&uid={uid}&key={key}"
sign = MD5(signStr).toUpperCase()
```

### 6.3 同步调度器（`internal/sync/syncer.go`）

```go
type Syncer struct {
    client    *quanyu.Client
    storage   *storage.MongoStorage
    provider  *device.Provider
    config    SyncConfig
    logger    *zap.Logger
    cron      *cron.Cron
}

func (s *Syncer) Start(ctx context.Context) error
func (s *Syncer) Stop()
```

**调度流程**：

```
main()
  ├── 初始化配置
  ├── 初始化日志
  ├── 初始化 MongoDB 连接
  ├── 初始化全裕 API 客户端
  ├── 初始化设备列表 Provider
  ├── 启动回调 HTTP 服务
  ├── 初始化同步调度器
  │   ├── 注册 cron 任务（每种数据类型一个）
  │   └── 启动订阅续订定时器
  ├── 注册优雅退出信号处理
  └── 等待退出信号
```

### 6.4 增量同步逻辑（核心改进）

**每个 uid 独立跟踪同步进度**：

```go
// 以 SyncHistoryData 为例的伪代码
func (s *Syncer) SyncHistoryData(ctx context.Context, uid string) error {
    // 1. 获取该 uid 上次同步时间点
    state, _ := s.storage.GetSyncState(ctx, uid, "history_data")
    startTime := state.LastSyncTime
    if startTime == "" {
        startTime = time.Now().Add(-config.DefaultRange).Format("2006-01-02 15:04:05")
    }
    endTime := time.Now().Format("2006-01-02 15:04:05")

    // 2. 分页拉取所有数据
    page := 1
    for {
        resp, err := s.client.GetBatteryData(ctx, uid, startTime, endTime, 0)
        if err != nil {
            // 记录错误但不影响其他 uid
            s.storage.UpdateSyncError(ctx, uid, "history_data", err)
            return err
        }

        // 3. 批量写入 MongoDB（upsert 去重）
        s.storage.UpsertBatteryHistory(ctx, uid, resp)

        // 4. 检查是否还有下一页
        if page >= resp.TotalPage {
            break
        }
        page++
    }

    // 5. 成功后立即更新该 uid 的时间点
    s.storage.UpdateSyncTime(ctx, uid, "history_data", endTime)
    return nil
}
```

**关键改进点**（对比现有 quanyu-sync）：
1. **每个 uid 独立更新时间点** — 不再是"全部成功才更新"
2. **分页遍历** — 不再只拉第一页
3. **MongoDB 唯一索引 upsert** — 即使重复拉取也不会产生重复数据
4. **错误隔离** — 单个 uid 失败不影响其他 uid

### 6.5 订阅管理（`internal/sync/subscriber.go`）

```go
type Subscriber struct {
    client   *quanyu.Client
    provider *device.Provider
    config   SubscribeConfig
    logger   *zap.Logger
}

func (s *Subscriber) StartRenewLoop(ctx context.Context)     // 启动续订循环
func (s *Subscriber) subscribeAll(ctx context.Context) error // 分批订阅所有设备
```

**订阅续订流程**：

```
每 25 分钟执行一次：
  1. 从 Provider 获取当前 uid 列表
  2. 按 batch_size=20 分批
  3. 对每批调用 subscribeV2，subData=["info","alarm","online"]
  4. notifyurl = config.notifyurl_base + "/callback/push"
  5. 记录订阅结果日志
```

### 6.6 回调处理（`internal/callback/handler.go`）

```go
type Handler struct {
    storage *storage.MongoStorage
    logger  *zap.Logger
}

// 路由设计
// POST /callback/push — 统一接收全裕推送
//   根据 data 内容判断类型（info/alarm/online），分发处理

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request)
func (h *Handler) handleInfoPush(ctx context.Context, payload *InfoPushPayload)
func (h *Handler) handleAlarmPush(ctx context.Context, payload *AlarmPushPayload)
func (h *Handler) handleOnlinePush(ctx context.Context, payload *OnlinePushPayload)
```

**回调数据模型**：

```go
// info 推送 — 从 Python BatteryPushPayload 参考
type InfoPushPayload struct {
    AppID string         `json:"appid"`
    Data  []InfoPushItem `json:"data"`
}

type InfoPushItem struct {
    UID       string `json:"uid"`
    DevType   int    `json:"devtype,omitempty"`
    SN        string `json:"sn,omitempty"`
    Loc       string `json:"loc,omitempty"`
    Remain    int    `json:"remain,omitempty"`
    Online    int    `json:"online,omitempty"`
    Voltage   int    `json:"voltage,omitempty"`
    Charge    int    `json:"charge,omitempty"`
    Discharge int    `json:"discharge,omitempty"`
    BatTime   string `json:"bat_time,omitempty"`
}

// alarm 推送
type AlarmPushPayload struct {
    AppID string          `json:"appid"`
    Data  []AlarmPushItem `json:"data"`
}

type AlarmPushItem struct {
    UID     string         `json:"uid"`
    InfoObj map[string]any `json:"infoObj,omitempty"`
    Events  []AlarmEvent   `json:"event"`
}

type AlarmEvent struct {
    Alarm string `json:"alarm"`
    Type  int    `json:"type"`  // 0-恢复 1-产生
    Time  string `json:"time"`
}

// online 推送 — 复用 InfoPushPayload 结构，通过 online 字段判断
```

**回调处理流程**：

```
收到 POST /callback/push
  ├── 解析 JSON body
  ├── 判断数据类型（info: 有 remain/voltage 字段; alarm: 有 event 字段; online: 仅有 online 字段）
  ├── 存入对应 MongoDB 集合
  └── 如果是 info 类型，同时更新 battery_details 快照
```

> **注意**：根据全裕 API 的 subscribeV2 设计，三种推送（info/alarm/online）可能使用同一个 notifyurl，通过数据内容区分类型。也可能分别使用不同的 notifyurl。具体实现时需根据全裕实际行为调整。初始设计统一一个 endpoint，根据数据特征分发。

### 6.7 设备列表 Provider（`internal/device/provider.go`）

```go
type Provider struct {
    apiURL     string
    httpClient *http.Client
    interval   time.Duration
    mu         sync.RWMutex
    uids       []string
    logger     *zap.Logger
}

func (p *Provider) Start(ctx context.Context) error  // 启动定时刷新
func (p *Provider) GetUIDs() []string                 // 获取当前 uid 列表（线程安全）
func (p *Provider) refreshUIDs(ctx context.Context) error
```

---

## 7. 日志设计

### 7.1 日志分类

| 日志类型 | 输出位置 | 内容 |
|----------|----------|------|
| 运行日志 | console + `logs/app.log` | 程序启动/停止、任务调度、错误信息 |
| API 调试日志 | `logs/debug/api.log` | 每次全裕 API 请求的完整 URL、参数、响应 |
| 同步统计日志 | `logs/sync.log` | 每轮同步的统计摘要 |

### 7.2 结构化日志字段

```go
// API 请求日志
logger.Info("API request",
    zap.String("endpoint", endpoint),
    zap.String("uid", uid),
    zap.Int("attempt", attempt),
    zap.Duration("elapsed", elapsed),
)

// 同步统计日志
logger.Info("sync summary",
    zap.String("type", "history_data"),
    zap.Int("total_uids", total),
    zap.Int("success", success),
    zap.Int("failed", failed),
    zap.Duration("duration", totalElapsed),
)

// 回调接收日志
logger.Info("callback received",
    zap.String("type", "info"),
    zap.Int("data_count", len(payload.Data)),
    zap.String("appid", payload.AppID),
)
```

---

## 8. 测试设计

### 8.1 单元测试

| 测试文件 | 覆盖内容 |
|----------|----------|
| `sign_test.go` | 签名算法正确性、边界情况（空字符串、特殊字符、中文） |
| `models_test.go` | API 响应 JSON 解析、可选字段处理、异常数据容错 |
| `config_test.go` | 配置文件加载、默认值、非法配置检测 |
| `syncer_test.go` | 增量同步逻辑、时间点计算、分页判断 |
| `callback_test.go` | 推送数据解析、类型判断、字段映射 |

### 8.2 集成测试

| 测试文件 | 覆盖内容 |
|----------|----------|
| `client_test.go` | 使用 httptest mock 全裕 API，测试重试逻辑、错误码处理、超时、分页遍历 |
| `storage_test.go` | 使用真实 MongoDB（testcontainers 或本地实例），测试 upsert 去重、索引约束、并发写入 |

### 8.3 Mock Server 设计

```go
// 模拟全裕 API 行为
func setupMockServer() *httptest.Server {
    mux := http.NewServeMux()

    // /sw/api/open/device/detail — 返回电池详情
    mux.HandleFunc("/sw/api/open/device/detail", func(w http.ResponseWriter, r *http.Request) {
        // 验证签名、返回 mock 数据
    })

    // /sw/api/open/device/data — 返回分页时序数据
    mux.HandleFunc("/sw/api/open/device/data", func(w http.ResponseWriter, r *http.Request) {
        // 根据 page 参数返回不同页的数据
    })

    // ... 其他端点

    return httptest.NewServer(mux)
}
```

### 8.4 测试覆盖率目标

| 模块 | 目标覆盖率 |
|------|-----------|
| `quanyu/sign` | > 95% |
| `quanyu/client` | > 80% |
| `quanyu/models` | > 90% |
| `storage/*` | > 75% |
| `sync/*` | > 70% |
| `callback/*` | > 80% |

---

## 9. 启动流程

```
┌─────────────────────────────────────────────────┐
│                   main.go                        │
├─────────────────────────────────────────────────┤
│ 1. 加载 config.yaml                              │
│ 2. 初始化 zap logger                             │
│ 3. 连接 MongoDB，确保索引存在                      │
│ 4. 创建全裕 API 客户端                            │
│ 5. 创建设备列表 Provider，首次加载 uid 列表         │
│ 6. 创建回调 Handler，启动 HTTP 服务                │
│ 7. 创建 Syncer，注册 cron 任务                    │
│ 8. 创建 Subscriber，启动订阅续订                   │
│ 9. 等待 OS 信号（SIGINT/SIGTERM）                  │
│    ↓ 收到信号                                     │
│ 10. 停止 cron 调度器                              │
│ 11. 停止订阅续订                                  │
│ 12. 关闭 HTTP 服务                                │
│ 13. 关闭 MongoDB 连接                             │
│ 14. 刷新日志缓冲                                  │
└─────────────────────────────────────────────────┘
```

---

## 10. 错误处理策略

| 场景 | 处理方式 |
|------|----------|
| 单个 uid API 调用失败 | 记录错误，更新 sync_state.error，继续处理下一个 uid |
| 全裕 API 整体不可达 | 触发重试机制（3次），全部失败后等待下一轮 cron |
| MongoDB 写入失败 | 记录错误日志，不更新 sync_state 时间点，下次重试 |
| MongoDB 唯一索引冲突 | 静默忽略（upsert 语义，重复数据自动跳过） |
| 设备列表 API 失败 | 保留上次缓存的 uid 列表继续工作，下一周期重试 |
| 订阅续订失败 | 记录警告日志，下一周期重试（不影响已生效的订阅） |
| 回调 HTTP 服务异常 | 记录错误日志，全裕会重试推送 |

---

## 11. 实现分期

### Phase 1: 基础框架
- 项目初始化（go mod、目录结构）
- 配置加载
- 日志系统
- 签名算法 + 单测

### Phase 2: API 客户端
- 全裕 API 客户端（含重试）
- 所有 API 方法实现
- 数据模型定义
- Mock server 集成测试

### Phase 3: MongoDB 存储
- MongoDB 连接管理
- 所有集合的 CRUD + 索引
- sync_state 管理
- MongoDB 集成测试

### Phase 4: 同步逻辑
- 设备列表 Provider
- 5 种数据类型的增量同步
- 分页遍历
- cron 调度器

### Phase 5: 订阅与回调
- subscribeV2 分批订阅 + 续订
- HTTP 回调服务
- 回调数据处理与存储

### Phase 6: 集成与完善
- main.go 启动流程整合
- 优雅退出
- 端到端测试
- 启动脚本（start.bat / start.sh）
