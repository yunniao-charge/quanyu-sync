package quanyu

import (
	"encoding/json"
)

// QuanyuResponse 全裕API通用响应格式
type QuanyuResponse struct {
	Errno  int             `json:"errno"`
	Errmsg string          `json:"errmsg"`
	Data   json.RawMessage `json:"data"`
}

// BatteryDetail 电池详情数据
type BatteryDetail struct {
	UID          string      `json:"uid,omitempty"`
	Date         string      `json:"date,omitempty"`
	SOC          int         `json:"soc,omitempty"`
	SOH          interface{} `json:"soh,omitempty"`
	Voltage      string      `json:"voltage,omitempty"`
	OnlineStatus interface{} `json:"onlineStatus,omitempty"`
	DevState     int         `json:"devstate,omitempty"`
	DevStateDesc int         `json:"devstatedscg,omitempty"`
	CellNum      int         `json:"cellNum,omitempty"`
	CellVoltage  interface{} `json:"cellVoltage,omitempty"`
	DeviceBMS1   interface{} `json:"deviceBMS1,omitempty"`
	DeviceBMS2   interface{} `json:"deviceBMS2,omitempty"`
	DeviceCT1    interface{} `json:"deviceCT1,omitempty"`
	DeviceCT2    interface{} `json:"deviceCT2,omitempty"`
	Current      string      `json:"currentCurrent,omitempty"`
	Remain       string      `json:"remain,omitempty"`
	Charge       int         `json:"charge,omitempty"`
	Discharge    int         `json:"discharge,omitempty"`
	LastPos      string      `json:"lastPos,omitempty"`
	LastAct      string      `json:"lastAct,omitempty"`
	LocTime      string      `json:"loctime,omitempty"`
	Signal       interface{} `json:"signal,omitempty"`
	RSSI         interface{} `json:"rssi,omitempty"`
	IMSI         string      `json:"imsi,omitempty"`
	IMEI         string      `json:"imei,omitempty"`
	Sate         interface{} `json:"sate,omitempty"`
	GPSCount     int         `json:"gpsCount,omitempty"`
	N            string      `json:"n,omitempty"`
	E            string      `json:"e,omitempty"`
	Distance     string      `json:"distance,omitempty"`
	DistanceADay string      `json:"distanceADay,omitempty"`
	EnergyLeft   int         `json:"energyLeft,omitempty"`
	EnergyTotal  int         `json:"energyTotal,omitempty"`
	BTCode       string      `json:"btCode,omitempty"`
	BMSVersion   string      `json:"bmsVersion,omitempty"`
	AppVersion   string      `json:"appVersion,omitempty"`
	MAC          string      `json:"mac,omitempty"`
	Loop         int         `json:"loop,omitempty"`
	Dipper       int         `json:"dipper,omitempty"`
	StateCode    string      `json:"stateCode,omitempty"`
	ExtTemp1     int         `json:"extTemp1,omitempty"`
	ExtTemp2     int         `json:"extTemp2,omitempty"`
	ExtTemp3     int         `json:"extTemp3,omitempty"`
	ExtTemp4     int         `json:"extTemp4,omitempty"`
	TempMOS      int         `json:"tempMOS,omitempty"`
	TempENV      int         `json:"tempENV,omitempty"`
	TempC1       int         `json:"tempC1,omitempty"`
	TempC2       int         `json:"tempC2,omitempty"`
	Cap          string      `json:"cap,omitempty"`
	DevType      int         `json:"devtype,omitempty"`
}

// BatteryDataResponse 电池包数据查询响应
type BatteryDataResponse struct {
	UID          string      `json:"serialnum,omitempty"`
	OnlineStatus interface{} `json:"onlineStatus,omitempty"`
	TimeScale    interface{} `json:"timeScale,omitempty"`
	LP           string      `json:"lp,omitempty"`
	Sate         interface{} `json:"sate,omitempty"`
	DeviceTV     interface{} `json:"deviceTV,omitempty"`
	DeviceCT1    interface{} `json:"deviceCT1,omitempty"`
	DeviceCT2    interface{} `json:"deviceCT2,omitempty"`
	DeviceBMS1   interface{} `json:"deviceBMS1,omitempty"`
	DeviceBMS2   interface{} `json:"deviceBMS2,omitempty"`
	DeviceAV     interface{} `json:"deviceAV,omitempty"`
	DeviceCG     interface{} `json:"deviceCG,omitempty"`
	DeviceCC     interface{} `json:"deviceCC,omitempty"`
	DeviceSA     interface{} `json:"deviceSA,omitempty"`
	DeviceREM    interface{} `json:"deviceREM,omitempty"`
	DeviceCOREV  interface{} `json:"deviceCOREV,omitempty"`
	B1           interface{} `json:"b1,omitempty"`
	B2           interface{} `json:"b2,omitempty"`
	IMSI         string      `json:"imsi,omitempty"`
	MID          string      `json:"mid,omitempty"`
	StateCode    string      `json:"statecode,omitempty"`
	AppVer       string      `json:"appver,omitempty"`
	VVV          string      `json:"vvv,omitempty"`
	Cap          string      `json:"cap,omitempty"`
	Rem          string      `json:"rem,omitempty"`
}

// TracePoint 轨迹点
type TracePoint struct {
	Loc     string `json:"loc"`
	LocTime string `json:"loc_time"`
	BDS     string `json:"bds,omitempty"`
	RSSI    string `json:"rssi,omitempty"`
	GPS     string `json:"gps,omitempty"`
}

// BatteryTraceResponse 电池轨迹查询响应
type BatteryTraceResponse struct {
	UID          string       `json:"uid"`
	Pages        int          `json:"Pages,omitempty"`
	OnlineStatus interface{}  `json:"onlineStatus,omitempty"`
	Online       int          `json:"online,omitempty"`
	Trace        []TracePoint `json:"trace"`
}

// BatteryEventItem 电池事件数据项
type BatteryEventItem struct {
	V1 string      `json:"v1,omitempty"`
	V2 string      `json:"v2,omitempty"`
	V4 string      `json:"v4,omitempty"`
	V6 interface{} `json:"v6,omitempty"`
	V7 interface{} `json:"v7,omitempty"`
	V8 string      `json:"v8,omitempty"`
	V9 string      `json:"v9,omitempty"`
}

// BatteryEventResponse 电池事件数据查询响应
type BatteryEventResponse struct {
	Total int                `json:"total,omitempty"`
	List  []BatteryEventItem `json:"list,omitempty"`
}

// ChargeRecordItem 充放电记录项
type ChargeRecordItem struct {
	DeviceID      string      `json:"deviceid,omitempty"`
	ChargeBegin   string      `json:"chargebegin,omitempty"`
	ChargeEnd     string      `json:"chargeend,omitempty"`
	BeginSOC      int         `json:"beginsoc,omitempty"`
	EndSOC        int         `json:"endsoc,omitempty"`
	ChargeDWh     interface{} `json:"chargedwh,omitempty"`
	ChargeDAh     int         `json:"chargedah,omitempty"`
	AccAh         int         `json:"accah,omitempty"`
	DriveMiles    int         `json:"drivemiles,omitempty"`
	SiteName      string      `json:"siteName,omitempty"`
	CompanyName   string      `json:"companyName,omitempty"`
	BillType      int         `json:"billType,omitempty"`
	ChargeBillID  string      `json:"chargebillId,omitempty"`
	MaxCurrent    int         `json:"maxcurrent,omitempty"`
	MinCurrent    interface{} `json:"mincurrent,omitempty"`
	NetState      int         `json:"netstate,omitempty"`
	PauseNum      int         `json:"pausenum,omitempty"`
	IsEnd         int         `json:"isend,omitempty"`
	BeginVoltage  interface{} `json:"beginvoltage,omitempty"`
	EndVoltage    interface{} `json:"endvoltage,omitempty"`
	BeginTime     interface{} `json:"begintime,omitempty"`
	EndTime       interface{} `json:"endtime,omitempty"`
	BeginStart    interface{} `json:"beginStart,omitempty"`
	BeginEnd      interface{} `json:"beginEnd,omitempty"`
	ChargedStandV interface{} `json:"chargedstandvolt,omitempty"`
	IDXAuto       int         `json:"idxauto,omitempty"`
}

// ChargeDataResponse 电池充放电记录响应
type ChargeDataResponse struct {
	CurrPage   int                `json:"currPage,omitempty"`
	PageSize   int                `json:"pageSize,omitempty"`
	TotalCount int                `json:"totalCount,omitempty"`
	TotalPage  int                `json:"totalPage,omitempty"`
	List       []ChargeRecordItem `json:"list,omitempty"`
}

// SubscribeRequest 订阅请求参数
type SubscribeRequest struct {
	AppID     string   `json:"appid"`
	UID       string   `json:"uid"`
	NonceStr  string   `json:"nonce_str"`
	Sign      string   `json:"sign"`
	Time      string   `json:"time"`
	List      []string `json:"list"`
	SubData   []string `json:"sub_data"`
	NotifyURL string   `json:"notify_url"`
}
