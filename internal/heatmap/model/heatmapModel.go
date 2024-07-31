package model

import (
	"encoding/json"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
)

type Coordinate struct {
	Lat float64
	Lng float64
}

type HeatmapPoint struct {
	Coordinates ewkb.Point `db:"coordinates"`
	//Coordinates Coordinate
	Dbm int32 `db:"dbm"`
}

func (h *HeatmapPoint) String() string {
	//var coordinates = Coordinate{h.Coordinates.Lat, h.Coordinates.Lng}
	//coordinates := []Coordinate{h.Coordinates.X(), h.Coordinates.Y()}
	coordinates := []float64{h.Coordinates.X(), h.Coordinates.Y()}
	jsonObject := make(map[string]interface{})
	jsonObject["dbm"] = h.Dbm
	jsonObject["coordinates"] = coordinates

	js, err := json.Marshal(jsonObject)
	if err != nil {
		return ""
	}

	return string(js)
}

func (h *HeatmapPoint) UnmarshalJSON(bytes []byte) error {
	var jsonData struct {
		Dbm         int32     `json:"dbm"`
		Coordinates []float64 `json:"coordinates"`
	}

	err := json.Unmarshal(bytes, &jsonData)
	if err != nil {
		return err
	}
	h.Dbm = jsonData.Dbm
	//h.Coordinates = ewkb.Point{Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{jsonData.Coordinates[0], jsonData.Coordinates[1]}).SetSRID(4326)}
	point := ewkb.Point{Point: geom.NewPoint(geom.XY).MustSetCoords([]float64{jsonData.Coordinates[0], jsonData.Coordinates[1]}).SetSRID(4326)}
	//h.Coordinates = Coordinate{point.X(), point.Y()}
	h.Coordinates = point
	return nil
}
