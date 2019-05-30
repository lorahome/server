package transport

import (
	"context"
)

type MockLoRaTransport struct {
	Ch      chan []byte
	Error   error
	History [][]byte
}

func NewMockLoRaTransport() *MockLoRaTransport {
	return &MockLoRaTransport{
		Ch: make(chan []byte),
	}
}

func (m *MockLoRaTransport) Run(context.Context) error {
	return m.Error
}

func (m *MockLoRaTransport) Receive() <-chan []byte {
	return m.Ch
}

func (m *MockLoRaTransport) SendPacket(packet []byte) error {
	m.History = append(m.History, packet)
	return m.Error
}
