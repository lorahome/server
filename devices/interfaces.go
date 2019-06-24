package devices

import (
	"context"
)

type DeviceCreateFunc func(cfg interface{}, caps *Capabilities) (Device, error)

type Device interface {
	GetId() uint64
	GetName() string
	GetClassName() string
	GetUrl() string

	Start(ctx context.Context) error
	ProcessMessage(packet []byte) error
}
