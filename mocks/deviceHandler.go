package mocks

import (
	"github.com/lorahome/server/transport"
)

type MockDeviceHandler struct {
	url, name             string
	ProcessMessageHistory [][]byte
	Error                 error
}

func NewMockDeviceHandler(name, url string) *MockDeviceHandler {
	return &MockDeviceHandler{
		url:  url,
		name: name,
	}
}

func (m *MockDeviceHandler) Name() string {
	return m.name
}

func (m *MockDeviceHandler) Url() string {
	return m.url
}

func (m *MockDeviceHandler) ProcessMessage(source transport.Transport, packet []byte) error {
	m.ProcessMessageHistory = append(m.ProcessMessageHistory, packet)
	return m.Error
}
