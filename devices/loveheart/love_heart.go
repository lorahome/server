package multisensor

import (
	"encoding/hex"
	"fmt"
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

	// Private
	keyBytes     []byte
	influxClient *influxdb.InfluxDB
	mqttClient   *mqtt.MqttClient
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
	LightAls       string
	LightWhite     string
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
	}
	err := mapstructure.Decode(cfg, dev)
	if err != nil {
		return nil, err
	}

	// Convert AES key into byte array
	dev.keyBytes, err = hex.DecodeString(dev.Key)

	return dev, err
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

	glog.Infof("Got update from '%s':", s.Name)

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
	timestamp := time.Now()

	// Print values
	temp := float32(ls.Temperature / 100)
	volts := float64(ls.VoltageMv) / 1000
	glog.Infof("\tTemperature %.1fC", temp)
	glog.Infof("\tHumidity %d%%", ls.Humidity)
	glog.Infof("\tBattery %.2fV", volts)
	glog.Infof("\tLight ALS %d", ls.LightAls)
	glog.Infof("\tLight White %d", ls.LightWhite)

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

	// Emit measurements
	return s.influxClient.Write(points)
}

func init() {
	devices.RegisterDeviceClass(Url, NewLoveHeart)
}
