package devices

import (
	"github.com/lorahome/server/transport"
)

type CreateFunc func(cfg interface{}) (Device, error)

type Device interface {
	GetId() uint64
	GetName() string
	GetClassName() string
	GetUrl() string
	ProcessMessage(source transport.Transport, packet []byte) error
}

type BaseDevice struct {
	Id        uint64
	Name      string
	ClassName string
	Url       string
}

func (s *BaseDevice) GetName() string {
	return s.Name
}

func (s *BaseDevice) GetClassName() string {
	return s.ClassName
}

func (s *BaseDevice) GetId() uint64 {
	return s.Id
}

func (s *BaseDevice) GetUrl() string {
	return s.Url
}
