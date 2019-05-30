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
