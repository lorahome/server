package mocks

import (
	"context"
)

type MockTransport struct {
	Ch      chan []byte
	Error   error
	History [][]byte
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		Ch: make(chan []byte),
	}
}

func (m *MockTransport) Run(context.Context) error {
	return m.Error
}

func (m *MockTransport) Receive() <-chan []byte {
	return m.Ch
}

func (m *MockTransport) Send(packet []byte) error {
	m.History = append(m.History, packet)
	return m.Error
}

type Transport interface {
	Run(context.Context) error
	Receive() <-chan []byte
	Send([]byte) error
}
