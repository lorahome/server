package multisensor

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	influxClient "github.com/influxdata/influxdb1-client/v2"
	"github.com/mitchellh/mapstructure"

	pb "github.com/lorahome/devices/go/proto/loveheart"
	"github.com/lorahome/server/db/influxdb"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/encoding"
	"github.com/lorahome/server/mqtt"
	"github.com/lorahome/server/transport"
)

const (
	Url       = "https://github.com/lorahome/protobufs/blob/master/proto/loveheart/loveheart.proto"
	ClassName = "LoveHeart"
)

// Device
type LoveHeart struct {
	// Public parameters (being saved into YAML)
	devices.BaseDevice `yaml:",inline" mapstructure:",squash"`
	InfluxDb           *influxDbConfig
	Mqtt               *mqttConfig
	Key                string
	SequenceSend       uint32
	SequenceRecv       uint32

	// Private
	keyBytes     []byte
	influxClient *influxdb.InfluxDB
	mqttClient   *mqtt.MqttClient
	transport    transport.LoRaTransport

	// Current motion / animation state
	lastMotionDetected time.Time
	animationEnabled   bool
}

type influxDbConfig struct {
	Database     string
	Measurements *measurementsConfig
}

type measurementsConfig struct {
	Temperature    string
	Humidity       string
	LightALS       string
	LightWhite     string
	BatteryVoltage string
	BatteryPercent string
	Charging       string
	Animation      string
}

type mqttConfig struct {
	Topics *mqttTopicsConfig
	Retain bool
	Qos    byte
}

type mqttTopicsConfig struct {
	Temperature    string
	Humidity       string
	BatteryVoltage string
	BatteryPercent string
	LightAls       string
	LightWhite     string

	// Control topics
	Animation string
	Motion    string
}

func NewLoveHeart(cfg interface{}, caps *devices.Capabilities) (devices.Device, error) {
	// Create instance and map config values into struct
	dev := &LoveHeart{
		BaseDevice: devices.BaseDevice{
			Url:       Url,
			ClassName: ClassName,
		},
		influxClient: caps.InfluxDb,
		mqttClient:   caps.Mqtt,
		transport:    caps.Udp,
	}
	err := mapstructure.Decode(cfg, dev)
	if err != nil {
		return nil, err
	}

	// Convert AES key into byte array
	dev.keyBytes, err = hex.DecodeString(dev.Key)

	return dev, err
}

func (s *LoveHeart) Start(ctx context.Context) error {
	// var err error
	// motionTopicsCh <-chan *mqtt.MqttMessage
	motionCh, err := s.mqttClient.Subscribe(s.Mqtt.Topics.Motion, 0)
	if err != nil {
		return err
	}
	animationCh, err := s.mqttClient.Subscribe(s.Mqtt.Topics.Animation, 0)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-motionCh:
				val, _ := strconv.Atoi(msg.Value)
				if val > 0 {
					s.lastMotionDetected = time.Now()
				}
			case msg := <-animationCh:
				val, _ := strconv.Atoi(msg.Value)
				s.animationEnabled = !(val == 0)
				glog.Infof("LoveHeart: global animation state -> %v", s.animationEnabled)
			case <-ctx.Done():
				return
			}
		}
	}()

	return err
}

func (s *LoveHeart) ProcessMessage(encrypted []byte) error {
	// Decrypt message
	decrypted, err := encoding.AESdecryptCBC(s.keyBytes, encrypted)
	if err != nil {
		return err
	}

	// Unpack multiSensor protobuf
	ls := &pb.LoveHeartStatus{}
	err = proto.Unmarshal(decrypted, ls)
	if err != nil {
		return err
	}

	glog.Infof("Got status from '%s'", s.Name)

	timestamp := time.Now()
	hour, _, _ := timestamp.Clock()

	// Send response back to device while it is listening to radio
	payload := make([]byte, 256)
	s.SequenceSend++
	// Play animation only when enabled globally, there is motion in living room
	// and only with 10:00 - 20:00 timeframe
	animation := false
	if hour >= 10 && hour < 20 &&
		time.Since(s.lastMotionDetected) < 15*time.Minute &&
		s.animationEnabled {

		animation = true
	}
	resp := &pb.LoveHeartStatusResponse{
		EnableAnimation: animation,
		Sequence:        s.SequenceSend,
		Magic:           0xddeeff,
	}
	glog.Info("Sending response:")
	glog.Infof("\tanimation: %v", animation)
	serializedResp, _ := proto.Marshal(resp)
	encryptedResp, _ := encoding.AESencryptCBC(s.keyBytes, serializedResp)
	binary.LittleEndian.PutUint64(payload, s.Id)
	totalLen := len(encryptedResp) + 8
	copy(payload[8:], encryptedResp)
	err = s.transport.Send(payload[:totalLen])
	if err != nil {
		return err
	}

	// Prepare InfluxDB measurement points
	points, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Precision: "s",
		Database:  s.InfluxDb.Database,
	})
	if err != nil {
		return err
	}
	tags := map[string]string{
		"device_id":  fmt.Sprintf("%d", s.Id),
		"name":       s.Name,
		"class_name": s.ClassName,
	}

	// Print values
	temp := float32(ls.Temperature) / 100
	volts := float64(ls.VoltageMv) / 1000
	glog.Info("Values:")
	glog.Infof("\tTemperature %.1fC", temp)
	glog.Infof("\tHumidity %d%%", ls.Humidity)
	glog.Infof("\tBattery %.2fV, %d%%", volts, ls.BatteryPercents)
	glog.Infof("\tLight ALS %d", ls.LightAls)
	glog.Infof("\tLight White %d", ls.LightWhite)
	glog.Infof("\tCharging %v", ls.Charging)
	glog.Infof("\tAnimation %v", ls.Animation)

	// Process temperature
	point, err := influxClient.NewPoint(
		s.InfluxDb.Measurements.Temperature,
		tags,
		map[string]interface{}{
			"c": temp,
		},
		timestamp,
	)
	if err != nil {
		return err
	}
	points.AddPoint(point)

	err = s.mqttClient.Publish(s.Mqtt.Topics.Temperature,
		fmt.Sprintf("%.1f", temp),
		s.Mqtt.Qos,
		s.Mqtt.Retain,
	)
	if err != nil {
		return err
	}

	// Humidity
	point, err = influxClient.NewPoint(
		s.InfluxDb.Measurements.Humidity,
		tags,
		map[string]interface{}{
			"value": float32(ls.Humidity),
		},
		timestamp,
	)
	if err != nil {
		return err
	}
	points.AddPoint(point)

	s.mqttClient.Publish(s.Mqtt.Topics.Humidity,
		fmt.Sprintf("%d", ls.Humidity),
		s.Mqtt.Qos,
		s.Mqtt.Retain,
	)
	if err != nil {
		return err
	}

	// Battery voltage
	point, err = influxClient.NewPoint(
		s.InfluxDb.Measurements.BatteryVoltage,
		tags,
		map[string]interface{}{
			"voltage": volts,
		},
		timestamp,
	)
	if err != nil {
		return err
	}
	points.AddPoint(point)

	s.mqttClient.Publish(s.Mqtt.Topics.BatteryVoltage,
		fmt.Sprintf("%.2f", volts),
		s.Mqtt.Qos,
		s.Mqtt.Retain,
	)
	if err != nil {
		return err
	}

	// Battery percentage
	point, err = influxClient.NewPoint(
		s.InfluxDb.Measurements.BatteryPercent,
		tags,
		map[string]interface{}{
			"percent": ls.BatteryPercents,
		},
		timestamp,
	)
	if err != nil {
		return err
	}
	points.AddPoint(point)

	s.mqttClient.Publish(s.Mqtt.Topics.BatteryPercent,
		fmt.Sprintf("%d", ls.BatteryPercents),
		s.Mqtt.Qos,
		s.Mqtt.Retain,
	)
	if err != nil {
		return err
	}

	// Light ALS / White
	point, err = influxClient.NewPoint(
		s.InfluxDb.Measurements.LightALS,
		tags,
		map[string]interface{}{
			"als": ls.LightAls,
		},
		timestamp,
	)
	if err != nil {
		return err
	}
	points.AddPoint(point)

	s.mqttClient.Publish(s.Mqtt.Topics.LightAls,
		fmt.Sprintf("%d", ls.LightAls),
		s.Mqtt.Qos,
		s.Mqtt.Retain,
	)
	if err != nil {
		return err
	}

	point, err = influxClient.NewPoint(
		s.InfluxDb.Measurements.LightWhite,
		tags,
		map[string]interface{}{
			"white": ls.LightWhite,
		},
		timestamp,
	)
	if err != nil {
		return err
	}
	points.AddPoint(point)

	s.mqttClient.Publish(s.Mqtt.Topics.LightWhite,
		fmt.Sprintf("%d", ls.LightWhite),
		s.Mqtt.Qos,
		s.Mqtt.Retain,
	)
	if err != nil {
		return err
	}

	// Charging / animation
	point, err = influxClient.NewPoint(
		s.InfluxDb.Measurements.Charging,
		tags,
		map[string]interface{}{
			"value": boolToInt(ls.Charging),
		},
		timestamp,
	)
	if err != nil {
		return err
	}
	points.AddPoint(point)

	point, err = influxClient.NewPoint(
		s.InfluxDb.Measurements.Animation,
		tags,
		map[string]interface{}{
			"value": boolToInt(ls.Animation),
		},
		timestamp,
	)
	if err != nil {
		return err
	}
	points.AddPoint(point)

	// Emit measurements
	return s.influxClient.Write(points)
}

func boolToInt(val bool) int {
	if val {
		return 1
	}
	return 0
}

func init() {
	devices.RegisterDeviceClass(Url, NewLoveHeart)
}
