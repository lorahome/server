package registry

import (
	"testing"

	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry(t *testing.T) {
	// Clear registered devices list
	deviceList = map[uint64]devices.Device{}
	// Set up one device class - mock
	deviceClasses = map[string]devices.DeviceCreateFunc{
		mocks.Url: mocks.NewMockDevice,
	}

	// Add Mock device
	cfg := map[interface{}]interface{}{
		"id":  123,
		"url": mocks.Url,
	}
	dev, err := RegisterDevice(cfg)
	require.NoError(t, err)
	assert.Equal(t, uint64(123), dev.GetId())
	assert.Equal(t, mocks.Url, dev.GetUrl())

	// Find device by id
	res := GetDeviceById(123)
	assert.Equal(t, dev, res)

	// Negative: non existing device
	res = GetDeviceById(1)
	assert.Nil(t, res)
}
