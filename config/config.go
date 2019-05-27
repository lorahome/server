package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Devices []Device
	Udp     Udp
}

type Device struct {
	Id   string
	Name string
	Key  string
	Url  string
}

type Udp struct {
	Listen        string
	MaxPacketSize int `yaml:"max_packet_size"`
}

// LoadConfigFromFile reads and parses YAML configuration from file
func LoadConfigFromFile(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	config := &Config{}
	err = yaml.UnmarshalStrict(data, config)

	return config, err
}
