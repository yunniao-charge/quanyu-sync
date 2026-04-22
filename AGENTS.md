# AGENTS.md — quanyu-sync

Go 1.22 service that syncs battery IoT data from Quanyu (全裕) API into MongoDB. Runs cron jobs + HTTP callback server. See `CLAUDE.md` for full details.

## Commands

```bash
cp config.yaml.example config.yaml    # Edit before first run
go run cmd/main.go                     # Starts sync + callback server :8888
go test ./...                          # Run tests
```

No linter/formatter config present. TDD workflow described in `.claude/skills/`.

## Architecture

```
cmd/main.go              # Entry point
internal/
├── config/              # YAML config loading
├── quanyu/              # Quanyu API client
├── storage/             # MongoDB operations
├── sync/                # Cron sync jobs
├── callback/            # HTTP callback handlers
├── device/              # Device UID fetching from server
└── logger/              # Zap logger setup
config.yaml              # Runtime config (gitignored)
```

## Key Patterns

### Quanyu API

- **Signature**: `MD5("appid={appid}&nonce_str={nonce_str}&uid={uid}&key={key}").toUpperCase()`
- **Field naming**: camelCase (`subData`, `notifyurl`)
- **Subscribe**: must send all 3 types together: `subData: ["info", "alarm", "online"]`
- **trace API**: `errno=500` means no data for the period (normal), handle as empty
- **Docs are approximate** — actual behavior verified by tests

### Callback Data Format

Callback body uses `type` field to discriminate. `data` is a **single object**, not array:

| Type | Key fields | Notes |
|------|-----------|-------|
| `info` | `uid`, `remain`, `online`, `voltage`, `charge`, `discharge`, `bat_time`, `loc` | `bat_time` = `YYYYMMDDHHMMSS` → convert to `YYYY-MM-DD HH:MM:SS` |
| `alarm` | `uid`, `alarmData`, `time`, `alarmCode` | `alarmData` is JSON string, store as-is |
| `online` | `uid`, `online`, `time` | `time` is ms Unix timestamp |

### MongoDB Collections

All collections use **upsert** pattern:

| Collection | Unique key | Mode |
|------------|-----------|------|
| battery_details | uid | Keep latest |
| battery_history | uid+timestamp | Dedup |
| battery_traces | uid+loc_time | Dedup |
| battery_events | uid+alarm+time | Dedup |
| charge_records | uid+idx_auto | Dedup |
| callback_alarms | uid+alarm+time | Dedup |
| callback_online | uid+time | Dedup |
| sync_states | uid+sync_type | Upsert |

## Config

`config.yaml` (from `config.yaml.example`):
- `quanyu.*` — API credentials and base URL
- `device_api.url` — Server endpoint for fetching device UIDs
- `mongodb.*` — MongoDB connection
- `callback.*` — Listen address + public notify URL
- `subscribe.*` — Batch size, renew interval
- `sync.*` — Cron schedules and enabled flags per sync type

## Deployment

```bash
bash scripts/deploy.sh       # Deploy
bash scripts/restart.sh      # Restart
bash scripts/stop.sh         # Stop
```
Service file: `scripts/quanyu-sync.service`
