package registry

import (
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"github.com/lorahome/server/config"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/devices/sensor/multisensor"
)

// device protobuf url -> implementation
var deviceHandlers = map[string]devices.DeviceHandler{}

// device_id -> device
var deviceList = map[uint64]*devices.Device{}

func RegisterDevice(cfg *config.Device) (*devices.Device, error) {
	id, err := strconv.ParseUint(cfg.Id, 0, 64)
	if err != nil {
		return nil, err
	}

	// Is device already registered?
	if _, ok := deviceList[id]; ok {
		return nil, fmt.Errorf("Device %s (0x%x) already registered", cfg.Name, cfg.Id)
	}

	// Find device handler
	handler, ok := deviceHandlers[cfg.Url]
	if !ok {
		return nil, fmt.Errorf("Unknown device handler '%s' for device %s (0x%x)", cfg.Url, cfg.Name, id)
	}

	// Add new device
	device := &devices.Device{
		Id:      id,
		Config:  cfg,
		Key:     []byte(cfg.Key),
		Handler: handler,
	}
	deviceList[id] = device
	glog.Infof("Added device '%s' (0x%x) handled by %s",
		device.Config.Name, device.Id, device.Handler.Name(),
	)
	return device, nil
}

func GetDeviceById(deviceId uint64) *devices.Device {
	if device, ok := deviceList[deviceId]; ok {
		return device
	}
	return nil
}

func RegisterDeviceHandler(dev devices.DeviceHandler) {
	deviceHandlers[dev.Url()] = dev
}

func init() {
	RegisterDeviceHandler(multisensor.NewMultiSensor())
}
