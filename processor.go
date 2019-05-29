package main

import (
	"encoding/binary"
	"fmt"

	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/registry"
)

func processPacket(caps devices.Capabilities, packet []byte) error {
	// Parse device id
	deviceId, err := parseDeviceId(packet)
	if err != nil {
		return err
	}
	// Lookup for device handler
	device := registry.GetDeviceById(deviceId)
	if device == nil {
		return fmt.Errorf("device with id 0x%x does not exist", deviceId)
	}
	// Call device handler to process packet
	return device.ProcessMessage(caps, packet[8:])
}

func parseDeviceId(packet []byte) (uint64, error) {
	if len(packet) < 8 {
		return 0, fmt.Errorf("packet too short (%d)", len(packet))
	}
	id := binary.LittleEndian.Uint64(packet)
	return id, nil
}
