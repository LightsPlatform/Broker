package group

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/LightsPlatform/vSensor/sensor"
	log "github.com/sirupsen/logrus"
)

// Sensor represents sensor identification and url
type Sensor struct {
	ID  string
	URL string
}

// Group represents group of sensors under
// single point management
type Group struct {
	ID      string
	Sensors []Sensor

	OnData func(data []sensor.Data) `json:"-"`

	end chan int
}

// New creates new group with given identification
// and empty array of sensors.
func New(id string) *Group {
	return &Group{
		ID:      id,
		Sensors: make([]Sensor, 0),

		end: make(chan int, 1),
	}
}

// Add adds new sensor into group
func (g *Group) Add(s Sensor) {
	g.Sensors = append(g.Sensors, s)
}

// Run runs data collection loop from list of sensors.
func (g *Group) Run() {
	t := time.Tick(1 * time.Second)
	for {
		select {
		case <-t:
			for _, s := range g.Sensors {
				resp, err := http.Get(fmt.Sprintf("%s/api/sensor/%s/data", s.URL, s.ID))
				if err != nil {
					log.Errorln(err)
				}

				data, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Errorln(err)
					continue
				}
				if err := resp.Body.Close(); err != nil {
					log.Errorln(err)
					continue
				}

				var d []sensor.Data
				if err := json.Unmarshal(data, &d); err != nil {
					log.Errorln(err)
					continue
				}

				log.Infoln(d)
				g.OnData(d)
			}
		case <-g.end:
			return
		}
	}
}

// Stop stops data collection loop
func (g *Group) Stop() {
	g.end <- 1
}
