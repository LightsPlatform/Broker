package group

// Sensor represents sensor identification
type Sensor string

// Group represents group of sensors under
// single point management
type Group struct {
	ID      string
	Sensors []Sensor
}

// New creates new group with given identification
// and empty array of sensors.
func New(id string) *Group {
	return &Group{
		ID:      id,
		Sensors: make([]Sensor, 0),
	}
}

// Add adds new sensor into group
func (g *Group) Add(s Sensor) {
	g.Sensors = append(g.Sensors, s)
}
