package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testUrl = "testUrl"
)

func TestRegistry(t *testing.T) {
	// Set up one device class - mock
	deviceClasses = map[string]DeviceCreateFunc{
		testUrl: NewMockDevice,
	}

	// Add Mock device
	deviceList = map[uint64]Device{}
	cfg := map[interface{}]interface{}{
		"id":  123,
		"url": testUrl,
	}
	dev, err := RegisterDevice(cfg)
	require.NoError(t, err)
	assert.Equal(t, uint64(123), dev.GetId())
	assert.Equal(t, testUrl, dev.GetUrl())

	// Find device by id
	res := GetDeviceById(123)
	assert.Equal(t, dev, res)

	// Negative: non existing device
	res = GetDeviceById(1)
	assert.Nil(t, res)
}
