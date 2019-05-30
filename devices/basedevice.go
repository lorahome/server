package devices

// BaseDevice partially implements common methods of Device interface
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
