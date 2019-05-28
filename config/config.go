package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config is top level configuration object
type Config struct {
	Devices []interface{}
	Udp     Udp
}

// Udp holds parameters for UDP server
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
	cfg := &Config{}
	err = yaml.UnmarshalStrict(data, cfg)

	return cfg, err
}

// SaveToFile saves cfg into file fn
func SaveToFile(fn string, cfg *Config) {
	data, _ := yaml.Marshal(cfg)
	ioutil.WriteFile(fn, data, 0644)
}
