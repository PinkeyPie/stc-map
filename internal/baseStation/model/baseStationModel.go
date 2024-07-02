package model

import (
	"encoding/json"
	"fmt"
	uuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"time"
)

type BaseStation struct {
	ID             uint64     `db:"id"`
	ElevationAngle int16      `db:"elevation_angle"`
	LacTac         int32      `db:"lac_tac"`
	Cid            int32      `db:"cid"`
	SectorNumber   int16      `db:"sector_number"`
	Azimuth        int16      `db:"azimuth"`
	Height         float32    `db:"height"`
	Power          int16      `db:"power"`
	UsingStart     time.Time  `db:"using_start"`
	UsingStop      *time.Time `db:"using_stop"`
	Address        string     `db:"address"`
	Coordinates    ewkb.Point `db:"coordinates"`
	RegionId       uuid.UUID  `db:"region"`
	Comment        *string    `db:"comment"`
	Operators      []Operator
	Region         Region
}

type Region struct {
	ID   uuid.UUID `db:"id"`
	Name string    `db:"name"`
}

type Operator struct {
	ID           uuid.UUID `db:"id"`
	Name         string    `db:"name"`
	Mcc          int16     `db:"mcc"`
	Mnc          int16     `db:"mnc"`
	BaseStations []BaseStation
}

type OperatorBs struct {
	operator uuid.UUID `db:"operator"`
	bs       uint64    `db:"bs"`
}

type Waypoint struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Geometry json.RawMessage `json:"geometry"`
}

func (bs *BaseStation) String() string {
	var mcc int16 = 0
	var mnc int16 = 0
	operator := "unknown"
	if len(bs.Operators) != 0 {
		mcc = bs.Operators[0].Mcc
		mnc = bs.Operators[0].Mnc
		operator = bs.Operators[0].Name
	}
	return fmt.Sprintf("BaseStation{mcc:%d, mnc:%d, operator:%s, coords: (%f, %f)", mcc, mnc, operator, bs.Coordinates.Point.X(), bs.Coordinates.Point.Y())
}

func (bs *BaseStation) UnmarshalJSON(b []byte) error {
	var tmp struct {
		//Operator string    `json:"operator"`
		//Mcc      int16     `json:"mcc"`
		//Mnc      int16     `json:"mnc"`
		Coords []float64 `json:"coords"`
	}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	point := geom.NewPoint(geom.XY).MustSetCoords([]float64{tmp.Coords[0], tmp.Coords[1]}).SetSRID(4326)
	bs.Coordinates = ewkb.Point{Point: point}

	return nil
}
