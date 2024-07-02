package baseStation

import (
	cluster "github.com/aliakseiz/gocluster"
	"simpleServer/internal/baseStation/model"
)

type PointInfo struct {
	Operator string  `json:"operator"`
	Mcc      int16   `json:"mcc"`
	Mnc      int16   `json:"mnc"`
	Radius   float32 `json:"radius"`
}

type Point struct {
	Type   string    `json:"type"`
	Coords []float64 `json:"coords"`
	Info   PointInfo `json:"info"`
}

type Cluster struct {
	Type       string    `json:"type"`
	Coords     []float64 `json:"coords"`
	PointCount int       `json:"pointCount"`
}

func NewBaseStationResponse(bs *model.BaseStation) *Point {
	point := Point{
		Type:   "Point",
		Coords: []float64{bs.Coordinates.X(), bs.Coordinates.Y()},
		Info: PointInfo{
			Radius: 1,
		},
	}
	if len(bs.Operators) != 0 {
		point.Info.Mcc = bs.Operators[0].Mcc
		point.Info.Mnc = bs.Operators[0].Mnc
		point.Info.Operator = bs.Operators[0].Name
	}

	return &point
}

func NewClusterResponse(points []cluster.Point) []interface{} {
	data := make([]interface{}, 0)

	for i := range points {
		if points[i].NumPoints == 1 {
			point := Point{
				Type:   "point",
				Coords: []float64{points[i].Y, points[i].X},
				Info:   PointInfo{Radius: 1},
			}
			data = append(data, point)
		} else {
			data = append(data, Cluster{
				Type:       "cluster",
				Coords:     []float64{points[i].Y, points[i].X},
				PointCount: points[i].NumPoints,
			})
		}
	}

	return data
}
