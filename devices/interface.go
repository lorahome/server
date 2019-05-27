package devices

import (
	"github.com/lorahome/server/config"
	"github.com/lorahome/server/transport"
)

type DeviceHandler interface {
	Name() string
	Url() string
	ProcessMessage(source transport.Transport, packet []byte) error
}

type Device struct {
	Config  *config.Device
	Id      uint64
	Key     []byte
	Handler DeviceHandler
}
