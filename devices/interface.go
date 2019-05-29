package devices

type DeviceCreateFunc func(cfg interface{}) (Device, error)

type Device interface {
	GetId() uint64
	GetName() string
	GetClassName() string
	GetUrl() string
	ProcessMessage(caps Capabilities, packet []byte) error
}

type Capabilities interface {
	// Send sends arbitrary (up to device) packet back to LoRa network
	// using all available transports
	SendPacket(packet []byte) error
	// InfluxDbWrite writes points (usually metrics) into InfluxDB
	// InfluxDbWrite()
}

// BaseDevice partially implements common methods of Device interface
type BaseDevice struct {
	Id        uint64
	Name      string
	ClassName string
	Url       string
}

func (s *BaseDevice) GetName() string {
	return s.Name
}

func (s *BaseDevice) GetClassName() string {
	return s.ClassName
}

func (s *BaseDevice) GetId() uint64 {
	return s.Id
}

func (s *BaseDevice) GetUrl() string {
	return s.Url
}
