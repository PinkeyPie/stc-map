package model

import (
	"encoding/json"
	uuid "github.com/gofrs/uuid"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"time"
)

type Post struct {
	Id          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Coordinates []GpsData
}

type GpsData struct {
	Id          uuid.UUID  `db:"id"`
	Coordinates ewkb.Point `db:"coordinates"`
	Time        time.Time  `db:"time"`
	Altitude    float32    `db:"altitude"`
	Speed       float32    `db:"speed"`
	Heading     float32    `db:"heading"`
	PostId      uuid.UUID  `db:"post_id"`
	Post        *Post
}

func (p *Post) String() string {
	jsonObject := make(map[string]interface{})
	jsonObject["id"] = p.Id
	jsonObject["name"] = p.Name
	coordinates := make(map[string]interface{})
	for _, coordinate := range p.Coordinates {
		coordinates["coordinates"] = []float64{coordinate.Coordinates.X(), coordinate.Coordinates.Y()}
		coordinates["time"] = coordinate.Time.Format("15:04")
	}
	js, err := json.Marshal(jsonObject)
	if err != nil {
		return ""
	}
	return string(js)
}

func (p *Post) UnmarshalJson(bytes []byte) error {
	var jsonData struct {
		Id          uuid.UUID `json:"id"`
		Name        string    `json:"name"`
		Coordinates []struct {
			Coordinates []float64 `json:"coordinates"`
			Time        time.Time `json:"time"`
		} `json:"coordinates"`
	}
	err := json.Unmarshal(bytes, &jsonData)
	if err != nil {
		return err
	}
	p.Name = jsonData.Name
	p.Id = jsonData.Id
	p.Coordinates = make([]GpsData, 0)
	for _, coordinate := range jsonData.Coordinates {
		gpsData := GpsData{
			Time:   coordinate.Time,
			PostId: p.Id,
			Post:   p,
		}
		id, _ := uuid.NewV4()
		gpsData.Id = id
		point := geom.NewPoint(geom.XY).MustSetCoords([]float64{coordinate.Coordinates[0], coordinate.Coordinates[1]}).SetSRID(4326)
		gpsData.Coordinates = ewkb.Point{Point: point}
		p.Coordinates = append(p.Coordinates, gpsData)
	}

	return nil
}
