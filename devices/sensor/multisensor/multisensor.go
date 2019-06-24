package multisensor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	pb "github.com/belyalov/OpenIOT-protobufs/go/sensor"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	influxClient "github.com/influxdata/influxdb1-client/v2"
	"github.com/mitchellh/mapstructure"

	"github.com/lorahome/server/db/influxdb"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/encoding"
	"github.com/lorahome/server/mqtt"
)

const (
	Url       = "https://github.com/lorahome/protobufs/blob/master/proto/sensor/multisensor.proto"
	ClassName = "MultiSensor"
)

// Device
type MultiSensor struct {
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
}

func NewMultiSensor(cfg interface{}, caps *devices.Capabilities) (devices.Device, error) {
	// Create instance and map config values into struct
	dev := &MultiSensor{
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

	// Validate / fix InfluxDB config
	if dev.InfluxDb != nil {
		idb := dev.InfluxDb
		if idb.Database == "" {
			if caps.InfluxDb.DefaultDatabase != "" {
				idb.Database = caps.InfluxDb.DefaultDatabase
			} else {
				return nil, errors.New("InfluxDB database name is required")
			}
		}
		if idb.Measurements == nil {
			idb.Measurements = &measurementsConfig{}
		}
		msr := idb.Measurements
		if msr.BatteryVoltage == "" {
			msr.BatteryVoltage = "battery_voltage"
		}
		if msr.Temperature == "" {
			msr.Temperature = "temperature"
		}
		if msr.Humidity == "" {
			msr.Humidity = "humidity"
		}
	}

	// Convert AES key into byte array
	dev.keyBytes, err = hex.DecodeString(dev.Key)

	return dev, err
}

func (m *MultiSensor) Start(ctx context.Context) error {
	return nil
}

func (s *MultiSensor) ProcessMessage(encrypted []byte) error {
	// Decrypt message
	decrypted, err := encoding.AESdecryptCBC(s.keyBytes, encrypted)
	if err != nil {
		return err
	}

	// Unpack multiSensor protobuf
	ms := &pb.MultiSensorStatus{}
	err = proto.Unmarshal(decrypted, ms)
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

	// Process temperature
	if ms.Temperature != nil {
		glog.Infof("\tTemperature %vC %vF", ms.Temperature.ValueC, ms.Temperature.ValueF)
		// Influx
		point, err := influxClient.NewPoint(
			s.InfluxDb.Measurements.Temperature,
			tags,
			map[string]interface{}{
				"c": ms.Temperature.ValueC,
				"f": ms.Temperature.ValueF,
			},
			timestamp,
		)
		if err != nil {
			return err
		}
		points.AddPoint(point)
		// MQTT
		if s.Mqtt.Topics.Temperature != "" {
			err = s.mqttClient.Publish(s.Mqtt.Topics.Temperature,
				fmt.Sprintf("%.1f", ms.Temperature.ValueC),
				s.Mqtt.Qos,
				s.Mqtt.Retain,
			)
			if err != nil {
				return err
			}
		}
	}
	// Humidity
	if ms.Humidity != nil {
		glog.Infof("\tHumidity %.0f%%", ms.Humidity.Value)
		// Influx
		point, err := influxClient.NewPoint(
			s.InfluxDb.Measurements.Humidity,
			tags,
			map[string]interface{}{
				"value": ms.Humidity.Value,
			},
			timestamp,
		)
		if err != nil {
			return err
		}
		points.AddPoint(point)
		// MQTT
		if s.Mqtt.Topics.Humidity != "" {
			s.mqttClient.Publish(s.Mqtt.Topics.Humidity,
				fmt.Sprintf("%.0f", ms.Humidity.Value),
				s.Mqtt.Qos,
				s.Mqtt.Retain,
			)
			if err != nil {
				return err
			}
		}
	}
	// Add battery voltage
	if ms.Battery != nil {
		volts := float64(ms.Battery.VoltageMv) / 1000
		glog.Infof("\tBattery %1.2fV", volts)
		point, err := influxClient.NewPoint(
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
		// MQTT
		if s.Mqtt.Topics.BatteryVoltage != "" {
			s.mqttClient.Publish(s.Mqtt.Topics.BatteryVoltage,
				fmt.Sprintf("%0.2f", volts),
				s.Mqtt.Qos,
				s.Mqtt.Retain,
			)
			if err != nil {
				return err
			}
		}
	}
	// Emit measurements
	s.influxClient.Write(points)

	return nil
}

func init() {
	devices.RegisterDeviceClass(Url, NewMultiSensor)
}
