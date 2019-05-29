package registry

import (
	"errors"
	"fmt"

	"github.com/golang/glog"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/devices/sensor/multisensor"
)

// Map of url -> device creator func
var deviceClasses = map[string]devices.DeviceCreateFunc{}

// device_id -> device
var deviceList = map[uint64]devices.Device{}

func RegisterDeviceClass(url string, dev devices.DeviceCreateFunc) {
	deviceClasses[url] = dev
}

func RegisterDevice(cfg interface{}) (devices.Device, error) {
	if cfg == nil {
		return nil, errors.New("Invalid config")
	}
	// Find device class by URL
	cfgMap := cfg.(map[interface{}]interface{})
	url, ok := cfgMap["url"].(string)
	if !ok {
		return nil, errors.New("Invalid config: 'url' not found")
	}
	// Create instance of device class
	creator, ok := deviceClasses[url]
	if !ok {
		return nil, fmt.Errorf("Unknown device class '%s'", url)
	}
	dev, err := creator(cfg)
	if err != nil {
		return nil, err
	}
	deviceList[dev.GetId()] = dev
	glog.Infof("Added %s device: %s (%d)", dev.GetClassName(), dev.GetName(), dev.GetId())

	return dev, err
}

func GetDeviceById(id uint64) devices.Device {
	if device, ok := deviceList[id]; ok {
		return device
	}
	return nil
}

func GetDevicesForConfigSave() []interface{} {
	ret := []interface{}{}
	for _, dev := range deviceList {
		ret = append(ret, dev)
	}
	return ret
}

func init() {
	// Register all known devices
	RegisterDeviceClass(multisensor.Url, multisensor.NewMultiSensor)
}
