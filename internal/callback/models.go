package callback

// InfoPushPayload info 推送数据
type InfoPushPayload struct {
	AppID string      `json:"appId"`
	Type  string      `json:"type"`
	Data  InfoPushItem `json:"data"`
}

// InfoPushItem info 推送数据项（单个对象）
type InfoPushItem struct {
	UID          string `json:"uid"`
	DevType      int    `json:"devtype,omitempty"`
	SN           string `json:"sn,omitempty"`
	Loc          string `json:"loc,omitempty"`
	Remain       int    `json:"remain,omitempty"`
	Online       int    `json:"online,omitempty"`
	OnlineStatus int    `json:"onlineStatus,omitempty"`
	Voltage      string `json:"voltage,omitempty"`
	Charge       int    `json:"charge,omitempty"`
	Discharge    int    `json:"discharge,omitempty"`
	BatTime      string `json:"bat_time,omitempty"`
	LocTime      string `json:"loc_time,omitempty"`
}

// AlarmPushItem alarm 推送数据项（已验证，2026-04-15）
// data 是单个对象，alarmData 是 JSON 字符串（不解析直接存储）
type AlarmPushItem struct {
	UID       string `json:"uid"`
	AlarmData string `json:"alarmData"`
	Time      string `json:"time"`
	AlarmCode string `json:"alarmCode"`
}

// OnlinePushItem online 推送数据项（已验证，2026-04-15）
// time 是毫秒级 Unix 时间戳
type OnlinePushItem struct {
	UID    string `json:"uid"`
	Online int    `json:"online"`
	Time   int64  `json:"time"`
}
