package devices

type DeviceCreateFunc func(cfg interface{}) (Device, error)

type Device interface {
	GetId() uint64
	GetName() string
	GetClassName() string
	GetUrl() string
	ProcessMessage(caps *Capabilities, packet []byte) error
}
