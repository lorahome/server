package multisensor

import (
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/transport"
)

type multiSensor struct {
}

func NewMultiSensor() devices.DeviceHandler {
	return &multiSensor{}
}

func (s *multiSensor) Name() string {
	return "MultiSensor"
}

func (s *multiSensor) Url() string {
	return "https://github.com/lorahome/protobufs/blob/master/proto/sensor/multisensor.proto"
}

func (s *multiSensor) ProcessMessage(source transport.Transport, packet []byte) error {
	return nil
}
