package vineyard

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidPlotName = errors.New("plot name is required")
	ErrPlotNotFound    = errors.New("plot not found")
)

type LatLng struct {
	Lat float64
	Lng float64
}

type Plot struct {
	id      string
	name    string
	variety string
	ha      float64
	age     int32
	status  string
	polygon []LatLng
}

func NewPlot(name, variety string, ha float64, age int32, status string, polygon []LatLng) (Plot, error) {
	if name == "" {
		return Plot{}, ErrInvalidPlotName
	}
	return Plot{
		id:      uuid.New().String(),
		name:    name,
		variety: variety,
		ha:      ha,
		age:     age,
		status:  status,
		polygon: polygon,
	}, nil
}

func ReconstitutePlot(id, name, variety string, ha float64, age int32, status string, polygon []LatLng) Plot {
	return Plot{id: id, name: name, variety: variety, ha: ha, age: age, status: status, polygon: polygon}
}

func (p *Plot) GetID() string      { return p.id }
func (p *Plot) GetName() string    { return p.name }
func (p *Plot) GetVariety() string { return p.variety }
func (p *Plot) GetHa() float64     { return p.ha }
func (p *Plot) GetAge() int32      { return p.age }
func (p *Plot) GetStatus() string  { return p.status }

func (p *Plot) GetPolygon() []LatLng {
	cp := make([]LatLng, len(p.polygon))
	copy(cp, p.polygon)
	return cp
}

func (p *Plot) Update(name, variety string, ha float64, age int32, status string, polygon []LatLng) error {
	if name == "" {
		return ErrInvalidPlotName
	}
	p.name = name
	p.variety = variety
	p.ha = ha
	p.age = age
	p.status = status
	p.polygon = polygon
	return nil
}
