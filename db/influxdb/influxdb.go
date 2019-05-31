package influxdb

import (
	"time"

	"github.com/golang/glog"
	influxClient "github.com/influxdata/influxdb1-client/v2"
	"github.com/mitchellh/mapstructure"
)

type InfluxDB struct {
	Config influxClient.HTTPConfig `mapstructure:",squash"`

	client  influxClient.Client
	enabled bool
}

func NewInfluxDB(cfg interface{}) (*InfluxDB, error) {
	db := &InfluxDB{}
	if cfg == nil {
		// Bypass mode - influxdb disabled
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

	duration, resp, err := db.client.Ping(1 * time.Second)
	glog.Infof("InfluxDB %s OK, received in %v", resp, duration)

	return db, err
}

func (db *InfluxDB) Write(points influxClient.BatchPoints) error {
	if db.enabled {
		return db.client.Write(points)
	}

	// Bypass mode
	return nil
}
