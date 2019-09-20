package led_strip

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"strconv"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/mapstructure"

	pb "github.com/lorahome/devices/go/proto/light"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/encoding"
	"github.com/lorahome/server/mqtt"
	"github.com/lorahome/server/transport"
)

const (
	Url       = "https://github.com/lorahome/devices/blob/master/proto/light/led_strip.proto"
	ClassName = "LedStrip"
)

// Device
type LedStrip struct {
	// Public parameters (being saved into YAML)
	devices.BaseDevice `yaml:",inline" mapstructure:",squash"`
	Mqtt               *mqttConfig
	Key                string

	// Private
	keyBytes   []byte
	mqttClient *mqtt.MqttClient
	transport  transport.LoRaTransport
}

type mqttConfig struct {
	Topics *mqttTopicsConfig
	Retain bool
	Qos    byte
}

type mqttTopicsConfig struct {
	Status  string
	Control string
}

func NewLedStrip(cfg interface{}, caps *devices.Capabilities) (devices.Device, error) {
	// Create instance and map config values into struct
	dev := &LedStrip{
		BaseDevice: devices.BaseDevice{
			Url:       Url,
			ClassName: ClassName,
		},
		mqttClient: caps.Mqtt,
		transport:  caps.Udp,
	}
	err := mapstructure.Decode(cfg, dev)
	if err != nil {
		return nil, err
	}

	// Convert AES key into byte array
	dev.keyBytes, err = hex.DecodeString(dev.Key)

	return dev, err
}

func (s *LedStrip) Start(ctx context.Context) error {
	controlCh, err := s.mqttClient.Subscribe(s.Mqtt.Topics.Control, 0)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-controlCh:
				val, err := strconv.Atoi(msg.Value)
				if err != nil {
					glog.Errorf("parse error for topic %s: %v", s.Mqtt.Topics.Control, err)
					continue
				}
				glog.Infof("%s: set light level to %v", s.Name, val)

				payload := make([]byte, 256)
				state := &pb.LedStripStatus{
					Channels: []uint32{uint32(val)},
				}
				serializedResp, _ := proto.Marshal(state)
				encryptedResp, _ := encoding.AESencryptCBC(s.keyBytes, serializedResp)
				binary.LittleEndian.PutUint64(payload, s.Id)
				totalLen := len(encryptedResp) + 8
				copy(payload[8:], encryptedResp)
				err = s.transport.Send(payload[:totalLen])
				if err != nil {
					glog.Errorf("unable to send: %v", err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (s *LedStrip) ProcessMessage(encrypted []byte) error {
	// Decrypt message
	glog.Infof("%v", encrypted)
	decrypted, err := encoding.AESdecryptCBC(s.keyBytes, encrypted)
	if err != nil {
		return err
	}

	glog.Infof("packet %v, sz %d", decrypted, len(decrypted))

	state := &pb.LedStripStatus{}
	err = proto.Unmarshal(decrypted, state)
	if err != nil {
		return err
	}

	glog.Infof("%s status: %v", s.Name, state.Channels)

	return nil
}

func init() {
	devices.RegisterDeviceClass(Url, NewLedStrip)
}
