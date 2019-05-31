package multisensor

import (
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
	Key                string

	// Private
	keyBytes  []byte
	capInflux *influxdb.InfluxDB
}

type influxDbConfig struct {
	Database     string
	Measurements *measurementsConfig
}

type measurementsConfig struct {
	Temperature string
	Humidity    string
	Battery     string
}

func NewMultiSensor(cfg interface{}, caps *devices.Capabilities) (devices.Device, error) {
	// Create instance and map config values into struct
	dev := &MultiSensor{
		BaseDevice: devices.BaseDevice{
			Url:       Url,
			ClassName: ClassName,
		},
		capInflux: caps.InfluxDb,
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
		if msr.Battery == "" {
			msr.Battery = "battery"
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

	// TODO: implement something
	glog.Infof("Got update from '%s':", s.Name)

	// Emit InfluxDB measurements
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
	// Temperature
	if ms.Temperature != nil {
		glog.Infof("\tTemperature %vC %vF", ms.Temperature.ValueC, ms.Temperature.ValueF)
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
	}
	// Humidity
	if ms.Humidity != nil {
		glog.Infof("\tHumidity %.2f%%", ms.Humidity.Value)
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
	}
	// Add battery voltage
	if ms.Battery != nil {
		volts := float64(ms.Battery.VoltageMv) / 1000
		glog.Infof("\tBattery %1.2fV", volts)
		point, err := influxClient.NewPoint(
			s.InfluxDb.Measurements.Battery,
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
	}
	// Emit measurements
	s.capInflux.Write(points)

	return nil
}

func init() {
	devices.RegisterDeviceClass(Url, NewMultiSensor)
}
