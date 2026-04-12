package device

// DeviceProvider 设备 UID 列表提供者接口
type DeviceProvider interface {
	GetUIDs() []string
}

// 编译时检查 *Provider 实现了 DeviceProvider 接口
var _ DeviceProvider = (*Provider)(nil)
