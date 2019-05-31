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

func RegisterDevice(url string, cfg interface{}, caps *Capabilities) (Device, error) {
	if url == "" {
		return nil, errors.New("URL should not be empty")
	}
	// Create instance of device class
	if createFunc, ok := deviceClasses[url]; ok {
		device, err := createFunc(cfg, caps)
		if err != nil {
			return nil, err
		}
		deviceList[device.GetId()] = device
		glog.Infof("Added %s device: %s (%d)",
			device.GetClassName(), device.GetName(), device.GetId())
		return device, nil
	} else {
		return nil, fmt.Errorf("Unknown device class '%s'", url)
	}
}

func GetDeviceById(id uint64) Device {
	if device, ok := deviceList[id]; ok {
		return device
	}
	return nil
}
