# PRD: 全裕电池数据同步服务 - 完善测试用例覆盖

## Problem Statement

项目当前只有 `internal/quanyu/sign_test.go` 一个测试文件，仅覆盖签名生成算法。核心业务逻辑（API 客户端、MongoDB 存储、同步调度、回调处理、设备管理）完全没有测试覆盖。这意味着：
- 无法验证 API 客户端对全裕 API 的调用是否正确
- 无法在不连接真实服务的情况下验证同步逻辑
- 代码重构或新增功能时缺乏安全网
- 配置变更可能引入难以发现的回归问题

## Solution

为项目建立分层的测试体系：

1. **集成测试**（`//go:build integration`）：验证核心模块与外部服务的交互
   - `quanyu/client`：调用真实全裕 API，验证接口契约
   - `storage`：连接真实 MongoDB，验证 CRUD 和索引去重
   - `callback`：使用 httptest + 真实 MongoDB，验证 HTTP 推送处理
   - `sync`：端到端验证同步流程

2. **单元测试**：验证业务逻辑的输入输出正确性
   - `sync`：使用 mock 依赖测试分页、增量同步、错误处理
   - `device/provider`：使用 mock HTTP 测试设备列表获取和缓存刷新
   - `config`：YAML 解析和默认值测试
   - `quanyu/sign`：（已有，保持不变）

3. **接口抽象**：为核心模块提取接口，使上层模块可以在单元测试中使用 mock
   - `QuanyuClient` 接口
   - `Storage` 接口（或按关注点拆分）
   - `DeviceProvider` 接口

## User Stories

### 集成测试 - API 客户端

1. 作为开发者，我想用真实凭据调用全裕 API 的 device/detail 接口并验证返回数据结构，以便确认接口契约有效
2. 作为开发者，我想用真实凭据调用全裕 API 的 device/data 接口并验证分页参数和时间范围参数传递正确，以便确认历史数据拉取逻辑正确
3. 作为开发者，我想用真实凭据调用全裕 API 的 battrace 接口并验证分页响应的字段映射，以便确认轨迹数据结构正确
4. 作为开发者，我想用真实凭据调用全裕 API 的 device/eventData 接口并验证事件数据的解析，以便确认事件同步基础可靠
5. 作为开发者，我想用真实凭据调用全裕 API 的 device/chargeData 接口并验证充放电记录的字段映射，以便确认充电记录同步基础可靠
6. 作为开发者，我想测试 API 返回错误码时的错误处理行为，以便确认错误不会导致 panic 或数据丢失
7. 作为开发者，我想测试 API 超时和网络错误时的重试行为，以便确认重试逻辑按预期工作
8. 作为开发者，我想测试无效 UID 的 API 响应处理，以便确认系统能优雅处理不存在的设备
9. 作为开发者，我想验证每次 API 请求中的签名计算是正确的，以便确认认证不会失败

### 集成测试 - MongoDB 存储

10. 作为开发者，我想测试 battery_details 集合的 upsert 操作（插入新文档 + 更新已有文档），以便确认快照数据正确存储
11. 作为开发者，我想测试 battery_history 集合的批量 upsert 和唯一索引去重，以便确认历史数据不会重复插入
12. 作为开发者，我想测试 battery_traces 集合的批量 upsert 和唯一索引去重，以便确认轨迹数据不会重复
13. 作为开发者，我想测试 battery_events 集合的批量 upsert 和唯一索引去重，以便确认事件数据不会重复
14. 作为开发者，我想测试 charge_records 集合的批量 upsert 和唯一索引去重，以便确认充放电记录不会重复
15. 作为开发者，我想测试 callback_info 集合的 upsert 操作，以便确认 info 回调数据正确存储
16. 作为开发者，我想测试 callback_alarms 集合的插入操作，以便确认告警回调数据正确存储
17. 作为开发者，我想测试 callback_online 集合的插入操作，以便确认上下线回调数据正确存储
18. 作为开发者，我想测试 sync_states 的读取和更新（含增量时间推进），以便确认增量同步状态管理正确
19. 作为开发者，我想测试 MongoDB 连接失败时的错误处理，以便确认系统不会 panic

### 集成测试 - HTTP 回调处理

20. 作为开发者，我想向回调接口发送 info 类型的推送数据并验证数据存入 callback_info 和 battery_details，以便确认 info 回调的完整处理流程
21. 作为开发者，我想向回调接口发送 alarm 类型的推送数据并验证数据存入 callback_alarms，以便确认告警推送处理正确
22. 作为开发者，我想向回调接口发送 online 类型的推送数据并验证数据存入 callback_online，以便确认上下线推送处理正确
23. 作为开发者，我想向回调接口发送无效 JSON 并验证返回 400 错误，以便确认异常输入处理正确
24. 作为开发者，我想向回调接口发送 GET 请求并验证返回 405，以便确认只接受 POST 请求
25. 作为开发者，我想向回调接口发送空 data 数组并验证优雅处理，以便确认边界情况不报错

### 集成测试 - 同步流程

26. 作为开发者，我想端到端执行一次 detail 同步（真实 API → 真实 MongoDB），以便验证完整的数据同步链路
27. 作为开发者，我想端到端执行一次 history_data 同步（含时间范围），以便验证增量同步的完整链路
28. 作为开发者，我想端到端执行一次带分页的同步（如 trace 或 event），以便验证分页遍历不遗漏数据
29. 作为开发者，我想验证同步完成后 sync_state 正确更新，以便确认增量同步机制可靠

### 单元测试 - 同步逻辑

30. 作为开发者，我想用 mock API 测试分页逻辑（模拟多页数据），以便确认所有页面的数据都被收集
31. 作为开发者，我想用 mock API 测试增量时间范围计算（从上次同步时间到现在），以便确认时间窗口正确
32. 作为开发者，我想用 mock API 测试单个 UID 同步失败不影响其他 UID，以便确认错误隔离机制
33. 作为开发者，我想用 mock API 测试同步状态的创建和更新，以便确认增量同步起点正确
34. 作为开发者，我想用 mock Storage 测试数据写入时的错误不会导致 sync_state 错误更新，以便确认数据一致性

### 单元测试 - 设备管理

35. 作为开发者，我想用 mock HTTP 测试设备 UID 列表的获取和解析，以便确认设备列表刷新逻辑正确
36. 作为开发者，我想测试设备列表缓存未刷新时返回旧数据，以便确认缓存行为正确
37. 作为开发者，我想测试设备 API 失败时不覆盖已有的 UID 列表，以便确认失败不影响正常工作

### 单元测试 - 配置

38. 作为开发者，我想测试正常 YAML 配置的加载和解析，以便确认所有字段正确映射
39. 作为开发者，我想测试缺失字段时使用默认值，以便确认默认值合理
40. 作为开发者，我想测试无效 YAML 文件的错误处理，以便确认配置加载失败有清晰的错误信息

### 接口抽象

41. 作为开发者，我想让 Syncer 依赖 QuanyuClient 接口而非具体 *Client，以便在单元测试中可以 mock API 调用
42. 作为开发者，我想让 Syncer 依赖 Storage 接口而非具体 *MongoStorage，以便在单元测试中可以 mock 数据库操作
43. 作为开发者，我想让 Syncer 依赖 DeviceProvider 接口而非具体 *Provider，以便在单元测试中可以控制设备列表

## Implementation Decisions

### 接口抽象

- **QuanyuClient 接口**：提取 `Client` 的 6 个公开方法（GetBatteryDetail、GetBatteryData、GetBatteryTrace、GetBatteryEvents、GetChargeRecords、SubscribeV2）为接口定义。接口放在 `internal/quanyu/` 包中，由现有 `Client` 结构体隐式实现
- **Storage 接口**：按同步和回调两个关注点拆分为 `SyncStorage` 和 `CallbackStorage` 接口。接口放在 `internal/storage/` 包中
  - `SyncStorage`：涵盖 detail/history/trace/event/charge 的 CRUD 以及 sync_state 管理
  - `CallbackStorage`：涵盖 callback_info/alarm/online 的 CRUD 以及 UpdateBatteryDetailFromCallback
- **DeviceProvider 接口**：提取 `Provider` 的 `GetUIDs()` 方法为接口。接口放在 `internal/device/` 包中
- Syncer、Handler、Subscriber 改为依赖接口，构造函数参数类型从具体类型改为接口类型

### 测试文件组织

- **单元测试**：放在对应包的同级目录（如 `internal/sync/syncer_test.go`），默认 `go test ./...` 即可运行
- **集成测试**：放在 `tests/integration/` 目录下，使用 `//go:build integration` build tag，通过 `go test -tags=integration ./...` 运行
- **测试辅助工具**：`tests/integration/helper.go` 提供共享的配置加载、MongoDB 连接、测试数据清理等功能

### 集成测试配置

- 使用现有 `config.yaml` 加载配置（appid、key、baseURL、MongoDB URI 等）
- 集成测试动态从设备 Provider 获取可用的测试 UID
- 每个集成测试用例执行后清理测试数据（使用带标记的 test_run_id 字段或直接清理新增数据）

### Mock 策略

- 单元测试中的 mock 实现采用简单的手动 mock struct（项目体量不大，不需要引入 mockgen 等代码生成工具）
- 每个 mock struct 只实现测试用例需要的方法，使用函数字段实现可配置的行为

## Testing Decisions

### 什么是好的测试

- **测试外部行为而非实现细节**：通过公开接口调用被测模块，验证输入→输出的正确性，不检查内部状态或调用顺序
- **独立性**：每个测试用例独立运行，不依赖其他测试的执行结果或执行顺序
- **确定性**：相同代码的测试结果必须一致，避免依赖时间、随机数等不确定因素
- **可清理**：集成测试在测试完成后清理所有产生的测试数据，保持数据库干净
- **有意义的信息**：测试名称和断言消息应清楚表达期望行为

### 需要测试的模块

| 模块 | 测试类型 | 测试文件位置 |
|------|---------|-------------|
| `quanyu/sign` | 单元测试 | `internal/quanyu/sign_test.go`（已有） |
| `quanyu/client` | 集成测试 | `tests/integration/quanyu/client_test.go` |
| `storage/*` | 集成测试 | `tests/integration/storage/*_test.go` |
| `sync/*` | 单元测试 | `internal/sync/*_test.go` |
| `sync/*` | 集成测试 | `tests/integration/sync/*_test.go` |
| `callback/*` | 集成测试 | `tests/integration/callback/handler_test.go` |
| `device/provider` | 单元测试 | `internal/device/provider_test.go` |
| `device/provider` | 集成测试 | `tests/integration/device/provider_test.go` |
| `config/*` | 单元测试 | `internal/config/config_test.go` |

### 测试优先级

1. **P0 - 接口抽象**：提取接口，使上层模块可测试
2. **P1 - API 客户端集成测试**：验证与全裕 API 的真实交互，这是整个系统的数据来源
3. **P2 - 存储层集成测试**：验证 MongoDB 的 CRUD 和去重逻辑
4. **P3 - 同步逻辑单元测试**：验证分页、增量、错误处理等核心业务逻辑
5. **P4 - 回调处理集成测试**：验证 HTTP 推送的接收和存储
6. **P5 - 其余模块测试**：device provider、config 等

## Out of Scope

- **logger 模块测试**：logger 是纯基础设施初始化，不包含业务逻辑，测试价值低
- **main.go 测试**：入口文件的组装逻辑通过各模块的独立测试间接覆盖
- **压力测试/性能测试**：不在本 PRD 范围内
- **订阅（SubscribeV2）的集成测试**：订阅操作会修改远端状态，不适合频繁测试，仅在单元测试中 mock
- **并发安全测试**：device provider 的并发安全不在本 PRD 范围
- **testcontainers**：暂不引入 Docker 依赖，直接使用配置文件中指定的 MongoDB 实例
- **代码覆盖率目标**：不设定具体的覆盖率百分比目标，关注关键路径的测试质量

## Further Notes

### 运行测试的命令

```bash
# 运行所有单元测试
go test ./...

# 运行集成测试（需要网络和 MongoDB）
go test -tags=integration ./tests/integration/...

# 运行所有测试
go test -tags=integration ./...

# 查看覆盖率
go test -tags=integration -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

### 实施建议

- 按优先级逐步实施，每个优先级完成后提交
- 集成测试首次编写时先验证通过，确保外部服务可用
- 接口抽象的改动需要保持向后兼容，现有调用方（main.go）的改动最小化
