package influxdb

import (
	"time"

	"github.com/golang/glog"
	influxClient "github.com/influxdata/influxdb1-client/v2"
	"github.com/mitchellh/mapstructure"
)

type InfluxDB struct {
	Config          influxClient.HTTPConfig `mapstructure:",squash"`
	DefaultDatabase string

	client  influxClient.Client
	enabled bool
}

type KV map[string]interface{}

func NewInfluxDB(cfg interface{}) (*InfluxDB, error) {
	db := &InfluxDB{}
	if cfg == nil {
		// Bypass mode - influxdb disabled
		glog.Info("InfluxDB is not enabled")
		return db, nil
	}

	// Map configuration into structure
	err := mapstructure.Decode(cfg, db)
	if err != nil {
		return nil, err
	}

	db.client, err = influxClient.NewHTTPClient(db.Config)
	if err == nil {
		db.enabled = true
	}

	_, resp, err := db.client.Ping(1 * time.Second)
	if err == nil {
		glog.Infof("InfluxDB started, version %s", resp)
	}

	return db, err
}

func (db *InfluxDB) Write(points influxClient.BatchPoints) error {
	if db.enabled {
		return db.client.Write(points)
	}

	// Bypass mode
	return nil
}
