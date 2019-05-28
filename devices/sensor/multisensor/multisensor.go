package multisensor

import (
	"encoding/hex"
	"github.com/belyalov/OpenIOT-protobufs/go/sensor"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/encoding"
	"github.com/lorahome/server/transport"
	"github.com/mitchellh/mapstructure"
)

const (
	Url       = "https://github.com/lorahome/protobufs/blob/master/proto/sensor/multisensor.proto"
	ClassName = "MultiSensor"
)

// Device
type MultiSensor struct {
	devices.BaseDevice `yaml:",inline" mapstructure:",squash"`

	Key      string
	keyBytes []byte
}

func NewMultiSensor(cfg interface{}) (devices.Device, error) {
	// Create instance and map config values into struct
	dev := &MultiSensor{}
	dev.Url = Url
	dev.ClassName = ClassName
	err := mapstructure.Decode(cfg, dev)
	if err != nil {
		return nil, err
	}
	// Convert AES key into byte array
	dev.keyBytes, err = hex.DecodeString(dev.Key)

	return dev, err
}

func (s *MultiSensor) ProcessMessage(source transport.Transport, encrypted []byte) error {
	// Decrypt message
	glog.Infof("%s: Got %d bytes message", s.Name, len(encrypted))
	decrypted, err := encoding.AESdecryptCBC(s.keyBytes, encrypted)
	if err != nil {
		return err
	}

	// Unpack multiSensor protobuf
	ms := &sensor.MultiSensorStatus{}
	err = proto.Unmarshal(decrypted, ms)
	if err != nil {
		return err
	}
	// TODO: implement something

	return nil
}
