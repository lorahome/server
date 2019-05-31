package devices

type DeviceCreateFunc func(cfg interface{}, caps *Capabilities) (Device, error)

type Device interface {
	GetId() uint64
	GetName() string
	GetClassName() string
	GetUrl() string

	ProcessMessage(packet []byte) error
}
