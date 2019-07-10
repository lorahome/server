package main

import (
	"encoding/binary"
	"fmt"

	"github.com/golang/glog"

	"github.com/lorahome/server/devices"
)

func processPacket(packet []byte) error {
	// Parse device id
	deviceId, err := parseDeviceId(packet)
	if err != nil {
		return err
	}

	glog.Infof("Got packet from %x", deviceId)

	// Lookup for device handler
	device := devices.GetDeviceById(deviceId)
	if device == nil {
		return fmt.Errorf("device 0x%x does not exist", deviceId)
	}
	// Call device handler to process packet
	return device.ProcessMessage(packet[8:])
}

func parseDeviceId(packet []byte) (uint64, error) {
	if len(packet) < 8 {
		return 0, fmt.Errorf("parseDeviceId: packet too short (%d)", len(packet))
	}
	id := binary.LittleEndian.Uint64(packet)
	return id, nil
}
