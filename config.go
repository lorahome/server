package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Config is top level configuration for all features
type Config struct {
	Devices  []interface{}
	InfluxDb interface{}
	Udp      interface{}
	Mqtt     interface{}
}

// ConfigLoadFromFile reads and parses YAML configuration from file
func ConfigLoadFromFile(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = yaml.UnmarshalStrict(data, cfg)

	return cfg, err
}

// ConfigSaveToFile saves cfg into file fn
func ConfigSaveToFile(fn string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	ioutil.WriteFile(fn, data, 0644)

	return nil
}
