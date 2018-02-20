package group

// Sensor represents sensor identification
type Sensor string

// Group represents group of sensors under
// single point management
type Group struct {
	id      string
	sensors []Sensor
}

// New creates new group with given identification
// and empty array of sensors.
func New(id string) *Group {
	return &Group{
		id:      id,
		sensors: make([]Sensor, 0),
	}
}
