package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Quanyu    QuanyuConfig    `yaml:"quanyu"`
	DeviceAPI DeviceAPIConfig `yaml:"device_api"`
	MongoDB   MongoDBConfig   `yaml:"mongodb"`
	Callback  CallbackConfig  `yaml:"callback"`
	Subscribe SubscribeConfig `yaml:"subscribe"`
	Log       LogConfig       `yaml:"log"`
	Sync      SyncConfig      `yaml:"sync"`
}

type QuanyuConfig struct {
	AppID      string        `yaml:"appid"`
	Key        string        `yaml:"key"`
	BaseURL    string        `yaml:"base_url"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
}

type DeviceAPIConfig struct {
	URL             string        `yaml:"url"`
	RefreshInterval time.Duration `yaml:"refresh_interval"`
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

type LogConfig struct {
	Dir         string `yaml:"dir"`
	Level       string `yaml:"level"`
	MaxSize     int    `yaml:"max_size"`
	MaxBackups  int    `yaml:"max_backups"`
	MaxAge      int    `yaml:"max_age"`
	APIDebugLog bool   `yaml:"api_debug_log"`
}

type SyncTaskConfig struct {
	Cron         string        `yaml:"cron"`
	Enabled      bool          `yaml:"enabled"`
	TimeRange    time.Duration `yaml:"time_range"`
	DefaultRange time.Duration `yaml:"default_range"`
}

type SyncConfig struct {
	Detail      SyncTaskConfig `yaml:"detail"`
	HistoryData SyncTaskConfig `yaml:"history_data"`
	Trace       SyncTaskConfig `yaml:"trace"`
	Event       SyncTaskConfig `yaml:"event"`
	ChargeData  SyncTaskConfig `yaml:"charge_data"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 设置默认值
	if cfg.Quanyu.BaseURL == "" {
		cfg.Quanyu.BaseURL = "https://iot.xtrunc.com"
	}
	if cfg.Quanyu.Timeout == 0 {
		cfg.Quanyu.Timeout = 30 * time.Second
	}
	if cfg.Quanyu.MaxRetries == 0 {
		cfg.Quanyu.MaxRetries = 3
	}
	if cfg.DeviceAPI.RefreshInterval == 0 {
		cfg.DeviceAPI.RefreshInterval = 6 * time.Hour
	}
	if cfg.MongoDB.URI == "" {
		cfg.MongoDB.URI = "mongodb://localhost:27017"
	}
	if cfg.MongoDB.Database == "" {
		cfg.MongoDB.Database = "quanyu_battery"
	}
	if cfg.Callback.ListenAddr == "" {
		cfg.Callback.ListenAddr = ":8888"
	}
	if cfg.Subscribe.BatchSize == 0 {
		cfg.Subscribe.BatchSize = 20
	}
	if cfg.Subscribe.RenewInterval == 0 {
		cfg.Subscribe.RenewInterval = 25 * time.Minute
	}
	if len(cfg.Subscribe.SubData) == 0 {
		cfg.Subscribe.SubData = []string{"info", "alarm", "online"}
	}
	if cfg.Log.Dir == "" {
		cfg.Log.Dir = "logs"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Log.MaxSize == 0 {
		cfg.Log.MaxSize = 100
	}
	if cfg.Log.MaxBackups == 0 {
		cfg.Log.MaxBackups = 30
	}
	if cfg.Log.MaxAge == 0 {
		cfg.Log.MaxAge = 30
	}

	setDefaultSyncTask(&cfg.Sync.Detail, "0 */2 * * * *")
	setDefaultSyncTask(&cfg.Sync.HistoryData, "0 */5 * * * *")
	setDefaultSyncTask(&cfg.Sync.Trace, "0 */5 * * * *")
	setDefaultSyncTask(&cfg.Sync.Event, "0 */5 * * * *")
	setDefaultSyncTask(&cfg.Sync.ChargeData, "0 */10 * * * *")

	if cfg.Sync.HistoryData.DefaultRange == 0 {
		cfg.Sync.HistoryData.DefaultRange = 1 * time.Hour
	}
	if cfg.Sync.Trace.DefaultRange == 0 {
		cfg.Sync.Trace.DefaultRange = 1 * time.Hour
	}
	if cfg.Sync.Event.DefaultRange == 0 {
		cfg.Sync.Event.DefaultRange = 1 * time.Hour
	}
	if cfg.Sync.ChargeData.DefaultRange == 0 {
		cfg.Sync.ChargeData.DefaultRange = 2 * time.Hour
	}

	return &cfg, nil
}

func setDefaultSyncTask(cfg *SyncTaskConfig, cron string) {
	if cfg.Cron == "" {
		cfg.Cron = cron
	}
}
