package multisensor

import (
	"github.com/golang/glog"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/transport"
	// "github.com/lorahome/server/encoding"
	"github.com/mitchellh/mapstructure"
)

const (
	Url       = "https://github.com/lorahome/protobufs/blob/master/proto/sensor/multisensor.proto"
	ClassName = "MultiSensor"
)

// Device
type MultiSensor struct {
	devices.BaseDevice `yaml:",inline" mapstructure:",squash"`

	Key  string
	Shit string

	keyBytes []byte
}

func NewMultiSensor(cfg interface{}) (devices.Device, error) {
	dev := &MultiSensor{}
	dev.Url = Url
	dev.ClassName = ClassName
	err := mapstructure.Decode(cfg, dev)
	dev.Shit = "shit!"
	return dev, err
}

func (s *MultiSensor) ProcessMessage(source transport.Transport, packet []byte) error {
	// decrypted, err := encoding.
	glog.Infof("Processing %v", packet)
	return nil
}
