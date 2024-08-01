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
	ID          uint64     `db:"id"`
	Address     string     `db:"address"`
	Coordinates ewkb.Point `db:"coordinates"`
	RegionId    uuid.UUID  `db:"region"`
	Comment     *string    `db:"comment"`
	BsInfo      []BsInfo
	Operators   []Operator
	Arfcn       []Arfcn
	Region      Region
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

type Arfcn struct {
	ID                  uuid.UUID `db:"id"`
	ArfcnNumber         int64     `db:"arfcn_number"`
	Uplink              float64   `db:"uplink"`
	Downlink            float64   `db:"downlink"`
	Bandwidth           float64   `db:"bandwidth"`
	Band                string    `db:"band"`
	Modulation          uuid.UUID `db:"modulation"`
	CellularNetworkType string
}

type CellularNetworkType struct {
	ID   uuid.UUID `db:"id"`
	Type string    `db:"type"`
}

type BsInfo struct {
	Arfcn          uuid.UUID  `db:"arfcn"`
	Bs             uint64     `db:"bs"`
	OperatorId     uuid.UUID  `db:"operator_id"`
	Cid            int32      `db:"cid"`
	LacTac         int32      `db:"lac_tac"`
	ElevationAngle int16      `db:"elevation_angle"`
	SectorNumber   int16      `db:"sector_number"`
	Azimuth        int16      `db:"azimuth"`
	Height         float32    `db:"height"`
	Power          int16      `db:"power"`
	UsingStart     time.Time  `db:"using_start"`
	UsingStop      *time.Time `db:"using_stop"`
	Comment        string     `db:"comment"`
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
