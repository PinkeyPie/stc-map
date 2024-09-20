package baseStation

import (
	cluster "github.com/aliakseiz/gocluster"
	"github.com/gofrs/uuid"
	"simpleServer/internal/baseStation/model"
)

type PointInfo struct {
	Id              uint64           `json:"id"`
	Coordinates     []float64        `json:"coordinates"`
	BsInfo          []model.BsInfo   `json:"bsInfo"`
	BsOperatorsInfo []model.Operator `json:"operatorsInfo"`
	BsArfcns        []model.Arfcn    `json:"bsArfcns"`
}

type OperatorsList struct {
	Operators []string `json:"operators"`
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
		Id: bs.ID,
	}

	point.Coordinates = []float64{bs.Coordinates.X(), bs.Coordinates.Y()}

	if len(bs.Operators) != 0 {
		point.BsOperatorsInfo = bs.Operators
	}

	if len(bs.BsInfo) != 0 {
		point.BsInfo = bs.BsInfo
	}

	if len(bs.Arfcn) != 0 {
		point.BsArfcns = bs.Arfcn
	}

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

func NewOperatorsResponse(operators []string) OperatorsList {
	operatorList := OperatorsList{}
	if len(operators) != 0 {
		operatorList.Operators = operators
	}
	return operatorList
}

func NewBsInfoResponse(bsInfo []model.BsInfo) []interface{} {
	data := make([]interface{}, 0)
	for i := range bsInfo {
		data = append(data, bsInfo[i])
	}
	return data
}
