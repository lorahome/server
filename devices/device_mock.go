package devices

import (
	"context"

	"github.com/mitchellh/mapstructure"
)

const (
	Url       = "testUrl"
	ClassName = "testClassName"
)

type MockDevice struct {
	BaseDevice `yaml:",inline" mapstructure:",squash"`

	ProcessMessageHistory [][]byte
	Error                 error
}

func NewMockDevice(cfg interface{}, _ *Capabilities) (Device, error) {
	ret := &MockDevice{}
	ret.Url = Url
	ret.ClassName = ClassName

	err := mapstructure.Decode(cfg, ret)

	return ret, err
}

func (m *MockDevice) Start(ctx context.Context) error {
	return nil
}

func (m *MockDevice) ProcessMessage(packet []byte) error {
	m.ProcessMessageHistory = append(m.ProcessMessageHistory, packet)
	return m.Error
}
