package registry

import (
	"testing"

	"github.com/lorahome/server/config"
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceHandler(t *testing.T) {
	// Clear device handler list
	deviceHandlers = map[string]devices.DeviceHandler{}
	// Add handler
	handler := mocks.NewMockDeviceHandler("test1", "http://test1")
	RegisterDeviceHandler(handler)
	assert.Equal(t, 1, len(deviceHandlers))
	// One more time - to be sure that old one will be overridden
	RegisterDeviceHandler(handler)
	assert.Equal(t, 1, len(deviceHandlers))
}

func TestDevice(t *testing.T) {
	// Clear device list
	deviceList = map[uint64]*devices.Device{}
	// Add mock device handler
	handler := mocks.NewMockDeviceHandler("test", "http://test")
	RegisterDeviceHandler(handler)

	// Add device
	cfg := &config.Device{
		Name: "test",
		Id:   "0x112233",
		Url:  "http://test",
		Key:  "11111111111111111111111111111111",
	}
	device, err := RegisterDevice(cfg)
	require.NotNil(t, device)
	assert.NoError(t, err)
	// Negative: add the same device
	_, err = RegisterDevice(cfg)
	assert.Error(t, err)

	// Try to find device handler by device id
	res := GetDeviceById(0x112233)
	require.NotNil(t, res)
	assert.Equal(t, device, res)
	// Negative: non existing device
	res = GetDeviceById(1)
	assert.Nil(t, res)
}
