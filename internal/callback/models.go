package callback

// InfoPushPayload info 推送数据
type InfoPushPayload struct {
	AppID string         `json:"appid"`
	Data  []InfoPushItem `json:"data"`
}

// InfoPushItem info 推送数据项
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

// AlarmPushPayload alarm 推送数据
type AlarmPushPayload struct {
	AppID string          `json:"appid"`
	Data  []AlarmPushItem `json:"data"`
}

// AlarmPushItem alarm 推送数据项
type AlarmPushItem struct {
	UID     string         `json:"uid"`
	InfoObj map[string]any `json:"infoObj,omitempty"`
	Events  []AlarmEvent   `json:"event"`
}

// AlarmEvent 告警事件
type AlarmEvent struct {
	Alarm string `json:"alarm"`
	Type  int    `json:"type"`
	Time  string `json:"time"`
}

// OnlinePushPayload online 推送数据
type OnlinePushPayload struct {
	AppID string          `json:"appid"`
	Data  []OnlinePushItem `json:"data"`
}

// OnlinePushItem online 推送数据项
type OnlinePushItem struct {
	UID    string `json:"uid"`
	Online int    `json:"online"`
	Time   string `json:"time,omitempty"`
}
