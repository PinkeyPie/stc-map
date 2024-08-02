package baseStation

import (
	cluster "github.com/aliakseiz/gocluster"
	"github.com/gofrs/uuid"
	"simpleServer/internal/baseStation/model"
)

type PointInfo struct {
	Id                  uint64 `json:"id"`
	LacTac              int32  `json:"lac_tac"`
	Cid                 int32  `json:"cid"`
	Operator            string `json:"operator"`
	Mcc                 int16  `json:"mcc"`
	Mnc                 int16  `json:"mnc"`
	ArfcnNumber         int64  `json:"arfcn_number"`
	CellularNetworkType string `json:"cellular_network_type"`
}

type ClusterViewPointInfo struct {
	ID int `json:"id"`
}

type Point struct {
	Type   string               `json:"type"`
	Coords []float64            `json:"coords"`
	Info   ClusterViewPointInfo ` json:"info"`
}

type Cluster struct {
	Type       string    `json:"type"`
	Coords     []float64 `json:"coords"`
	PointCount int       `json:"pointCount"`
}

type BsIdResponse struct {
	Id uuid.UUID `json:"id"`
}

func NewBaseStationResponse(bs *model.BaseStation) *PointInfo {
	point := PointInfo{
		Id:                  bs.ID,
		LacTac:              bs.BsInfo[0].LacTac,
		Cid:                 bs.BsInfo[0].Cid,
		Operator:            bs.Operators[0].Name,
		Mcc:                 bs.Operators[0].Mcc,
		Mnc:                 bs.Operators[0].Mnc,
		ArfcnNumber:         0,
		CellularNetworkType: "4G or something",
	}
	//if len(bs.Operators) != 0 {
	//	point.Info.Mcc = bs.Operators[0].Mcc
	//	point.Info.Mnc = bs.Operators[0].Mnc
	//	point.Info.Operator = bs.Operators[0].Name
	//}
	//if len(bs.Arfcn) != 0 {
	//	point.Info.ArfcnNumber = bs.Arfcn[0].ArfcnNumber
	//	point.Info.CellularNetworkType = bs.Arfcn[0].CellularNetworkTypes[0].Type
	//}
	return &point
}

func NewClusterResponse(points []cluster.Point) []interface{} {
	data := make([]interface{}, 0)

	for i := range points {
		if points[i].NumPoints == 1 {
			point := Point{
				Type: "point",
				Info: ClusterViewPointInfo{
					ID: points[i].ID + 1,
				},
				Coords: []float64{points[i].X, points[i].Y},
			}
			data = append(data, point)
		} else {
			data = append(data, Cluster{
				Type:       "cluster",
				Coords:     []float64{points[i].X, points[i].Y},
				PointCount: points[i].NumPoints,
			})
		}
	}
	return data
}
