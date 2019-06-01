package devices

import (
	"github.com/lorahome/server/db/influxdb"
	"github.com/lorahome/server/mqtt"
	"github.com/lorahome/server/transport"
)

type Capabilities struct {
	Udp      transport.LoRaTransport
	InfluxDb *influxdb.InfluxDB
	Mqtt     *mqtt.MqttClient
}
