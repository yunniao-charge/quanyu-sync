# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

全裕电池数据同步服务 - Go 语言实现，定时从全裕 API 拉取数据并接收推送回调，写入 MongoDB。

## 全裕 API 回调

### 回调数据格式

回调通过 `type` 字段区分，`data` 是单个对象（不是数组）。

#### info 回调

```json
{"type": "info", "data": {"uid": "xxx", "remain": 94, "online": 1, "voltage": "63249", "charge": 0, "discharge": 0, "bat_time": "20260415095539", "loc": "N31.66,E120.73"}}
```

- `bat_time` 格式为 `YYYYMMDDHHMMSS`，handler 统一转为 `YYYY-MM-DD HH:MM:SS`
- handler 直接更新 `battery_details` 的 soc/online_status/charge/discharge/loc/bat_time

#### alarm 回调

```json
{"type": "alarm", "data": {"uid": "xxx", "alarmData": "{...}", "time": "20260415111208", "alarmCode": "COVR"}}
```

- `alarmData` 是 JSON 字符串，直接存储不解析

#### online 回调

```json
{"type": "online", "data": {"uid": "xxx", "online": 0, "time": 1776219921000}}
```

- `time` 是毫秒级 Unix 时间戳

### API 注意事项

- 签名：`MD5("appid={appid}&nonce_str={nonce_str}&uid={uid}&key={key}").toUpperCase()`
- 字段名用驼峰：`subData`、`notifyurl`
- 订阅三种类型必须一起传：`subData: ["info", "alarm", "online"]`
- trace API 返回 errno=500 表示该时段无轨迹（正常），客户端已处理为空结果
- 文档仅供参考，实际行为以测试为准

## 数据存储

| 集合 | 模式 | 唯一键 |
|------|------|--------|
| battery_details | Upsert 最新 | uid |
| battery_history | Upsert 去重 | uid+timestamp |
| battery_traces | Upsert 去重 | uid+loc_time |
| battery_events | Upsert 去重 | uid+alarm+time |
| charge_records | Upsert 去重 | uid+idx_auto |
| callback_alarms | Upsert 去重 | uid+alarm+time |
| callback_online | Upsert 去重 | uid+time |
| sync_states | Upsert | uid+sync_type |

## 开发流程

使用 Skills 驱动的 TDD 流程。详见 `.claude/skills/` 和 `.claude/templates/`。
