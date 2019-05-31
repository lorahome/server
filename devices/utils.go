package devices

import (
	"io/ioutil"
	"strings"

	"github.com/golang/glog"
	"gopkg.in/yaml.v2"
)

// LoadFromFile reads and parses device definitions from YAML
func LoadFromFile(filename string, caps *Capabilities) error {
	// Read file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			// No file - no devices
			return nil
		}
		return err
	}
	// Parse YAML
	devices := map[string][]interface{}{}
	err = yaml.Unmarshal(data, devices)
	if err != nil {
		return err
	}
	// Clear all existing devices
	deviceList = map[uint64]Device{}
	// Register all devices from yaml
	for url, list := range devices {
		for _, device := range list {
			_, err := RegisterDevice(url, device, caps)
			if err != nil {
				glog.Fatalf("Unable to register device: %v", err)
			}
		}
	}

	return nil
}

// SaveToFile saves device list into filename
func SaveToFile(filename string) error {
	// Split all registered devices by url,
	// effectively making map of arrays
	toSave := map[string][]Device{}
	for _, dev := range deviceList {
		url := dev.GetUrl()
		toSave[url] = append(toSave[url], dev)
	}

	data, err := yaml.Marshal(toSave)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, data, 0644)
}
