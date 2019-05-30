package devices

import (
	"errors"
	"fmt"

	"github.com/golang/glog"
)

// Map of url -> device creator func
var deviceClasses = map[string]DeviceCreateFunc{}

// deviceId -> device
var deviceList = map[uint64]Device{}

func RegisterDeviceClass(url string, dev DeviceCreateFunc) {
	deviceClasses[url] = dev
}

func RegisterDevice(cfg interface{}) (Device, error) {
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

func GetDeviceById(id uint64) Device {
	if device, ok := deviceList[id]; ok {
		return device
	}
	return nil
}
