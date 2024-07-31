package heatmap

import (
	"simpleServer/internal/heatmap/model"
)

type Point struct {
	Dbm    int32     `json:"dbm"`
	Coords []float64 `'json:"coords"`
}

type HeatmapResponse struct {
	Dbm         int32     `json:"dbm"`
	Coordinates []float64 `json:"coordinates"`
}

func NewHeatmapPointsInBboxResponse(points []model.HeatmapPoint) []interface{} {
	data := make([]interface{}, 0)
	for _, point := range points {
		data = append(data, Point{
			Dbm:    point.Dbm,
			Coords: []float64{point.Coordinates.X(), point.Coordinates.Y()},
		})
	}
	return data
}
