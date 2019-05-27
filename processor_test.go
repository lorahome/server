package main

import (
	"testing"

	"github.com/lorahome/server/config"
	"github.com/lorahome/server/mocks"
	"github.com/lorahome/server/registry"
	"github.com/stretchr/testify/assert"
)

func TestParseId(t *testing.T) {
	runs := map[uint64][]byte{
		0x3531383105473831: {0x31, 0x38, 0x47, 0x05, 0x31, 0x38, 0x31, 0x35},
		0x0:                {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		0xFF00000000000000: {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
	}
	for id, packet := range runs {
		res, err := parseDeviceId(packet)
		assert.NoError(t, err)
		assert.Equal(t, id, res)
	}
	// Negative: too short
	short := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	_, err := parseDeviceId(short)
	assert.Error(t, err)
}

func TestProcessPacket(t *testing.T) {
	// End to end test of packet processing:
	// Setup device / device handler
	handler := mocks.NewMockDeviceHandler("packets", "packets.com")
	registry.RegisterDeviceHandler(handler)
	cfg := &config.Device{
		Name: "packets",
		Id:   "0x1234",
		Url:  "packets.com",
		Key:  "11111111111111111111111111111111",
	}
	registry.RegisterDevice(cfg)

	// "Send" packet
	source := mocks.NewMockTransport()
	packet := []byte{0x34, 0x12, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // id
		0x00, 0x01, 0x02, 0x03, // some payload
	}
	err := processPacket(source, packet)
	// Ensure that handler receives packet without device id
	assert.NoError(t, err)
	assert.Equal(t, []byte{0x00, 0x01, 0x02, 0x03}, handler.ProcessMessageHistory[0])

	// Negative: non existing device
	err = processPacket(source, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	assert.Error(t, err)
}
