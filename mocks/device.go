package mocks

import (
	"github.com/lorahome/server/devices"
	"github.com/lorahome/server/transport"
	"github.com/mitchellh/mapstructure"
)

const (
	Url       = "testUrl"
	ClassName = "testClassName"
)

type MockDevice struct {
	devices.BaseDevice `yaml:",inline" mapstructure:",squash"`

	ProcessMessageHistory [][]byte
	Error                 error
}

func NewMockDevice(cfg interface{}) (devices.Device, error) {
	ret := &MockDevice{}
	ret.Url = Url
	ret.ClassName = ClassName

	err := mapstructure.Decode(cfg, ret)

	return ret, err
}

func (m *MockDevice) ProcessMessage(source transport.Transport, packet []byte) error {
	m.ProcessMessageHistory = append(m.ProcessMessageHistory, packet)
	return m.Error
}
