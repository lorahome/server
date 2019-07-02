package multisensor

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
	influxClient "github.com/influxdata/influxdb1-client/v2"
	pb "github.com/lorahome/devices/go/proto/sensor"
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
	AmbientLight   string
	BatteryVoltage string
}

type mqttConfig struct {
	Topics        *mqttTopicsConfig
	Retain        bool
	Qos           byte
	ImperialUnits bool
}

type mqttTopicsConfig struct {
	Temperature       string
	Humidity          string
	AmbientLight      string
	AmbientLightWhite string
	BatteryVoltage    string
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

	// Prepare InfluxDB points / MQTT topics
	influxPoints := map[string]influxdb.KV{}
	mqttTopics := map[string]string{}
	if ms.Temperature != nil {
		if s.InfluxDb.Measurements.Temperature != "" {
			influxPoints[s.InfluxDb.Measurements.Temperature] = influxdb.KV{
				"c": ms.Temperature.ValueC,
				"f": ms.Temperature.ValueF,
			}
		}
		if s.Mqtt.Topics.Temperature != "" {
			if s.Mqtt.ImperialUnits {
				mqttTopics[s.Mqtt.Topics.Temperature] = fmt.Sprintf("%.1f", ms.Temperature.ValueF)
			} else {
				mqttTopics[s.Mqtt.Topics.Temperature] = fmt.Sprintf("%.1f", ms.Temperature.ValueC)
			}
		}
		glog.Infof("\tTemperature %vC, %vF", ms.Temperature.ValueC, ms.Temperature.ValueF)
	}
	if ms.Humidity != nil {
		if s.InfluxDb.Measurements.Humidity != "" {
			influxPoints[s.InfluxDb.Measurements.Humidity] = influxdb.KV{
				"value": ms.Humidity.Value,
			}
		}
		if s.Mqtt.Topics.Humidity != "" {
			mqttTopics[s.Mqtt.Topics.Humidity] = fmt.Sprintf("%.0f", ms.Humidity.Value)
		}
		glog.Infof("\tHumidity %v", ms.Humidity.Value)
	}
	if ms.AmbientLight != nil {
		if s.InfluxDb.Measurements.AmbientLight != "" {
			influxPoints[s.InfluxDb.Measurements.AmbientLight] = influxdb.KV{
				"als":   ms.AmbientLight.Value,
				"white": ms.AmbientLight.WhiteValue,
			}
		}
		if s.Mqtt.Topics.AmbientLight != "" {
			mqttTopics[s.Mqtt.Topics.AmbientLight] = fmt.Sprintf("%.0f", ms.AmbientLight.Value)
		}
		if s.Mqtt.Topics.AmbientLightWhite != "" {
			mqttTopics[s.Mqtt.Topics.AmbientLightWhite] = fmt.Sprintf("%.0f", ms.AmbientLight.WhiteValue)
		}
		glog.Infof("\tAmbientLight %v (white %v)", ms.AmbientLight.Value, ms.AmbientLight.WhiteValue)
	}
	if ms.Battery != nil {
		volts := float64(ms.Battery.VoltageMv) / 1000
		if s.InfluxDb.Measurements.BatteryVoltage != "" {
			influxPoints[s.InfluxDb.Measurements.BatteryVoltage] = influxdb.KV{
				"voltage": volts,
			}
		}
		if s.Mqtt.Topics.BatteryVoltage != "" {
			mqttTopics[s.Mqtt.Topics.BatteryVoltage] = fmt.Sprintf("%0.2f", volts)
		}
		glog.Infof("\tBattery Voltage %v", volts)
	}

	// Emit all InfluxDB points
	batchPoints, err := influxClient.NewBatchPoints(influxClient.BatchPointsConfig{
		Precision: "s",
		Database:  s.InfluxDb.Database,
	})
	if err != nil {
		return err
	}
	influxTags := map[string]string{
		"device_id":  fmt.Sprintf("%d", s.Id),
		"class_name": s.ClassName,
		"name":       s.Name,
	}
	timestamp := time.Now()
	for measurement, points := range influxPoints {
		point, err := influxClient.NewPoint(
			measurement,
			influxTags,
			points,
			timestamp,
		)
		if err != nil {
			return err
		}
		batchPoints.AddPoint(point)
	}
	err = s.influxClient.Write(batchPoints)
	if err != nil {
		return err
	}

	// Publish all MQTT topics
	for topic, value := range mqttTopics {
		s.mqttClient.Publish(topic, value, s.Mqtt.Qos, s.Mqtt.Retain)
		if err != nil {
			glog.Errorf("MQTT Publish failed: %v", err)
		}
	}

	return nil
}

func init() {
	devices.RegisterDeviceClass(Url, NewMultiSensor)
}
