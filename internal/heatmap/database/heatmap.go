package database

import (
	"context"
	"github.com/jmoiron/sqlx"
	"simpleServer/dbutils"
	"simpleServer/internal/heatmap/model"
	"simpleServer/pkg/logging"
)

type HeatmapDB interface {
	GetAllHeatmapPointsInBbox(ctx context.Context, n float64, w float64, s float64, e float64) ([]model.HeatmapPoint, error)
	GetAllHeatmapPointsByCoordsDB(ctx context.Context, lat float64, Lng float64) ([]model.HeatmapPoint, error)
	GetHeatmapPointsByIdDB(ctx context.Context, id int) ([]model.HeatmapPoint, error)
}

type heatmapDB struct {
	dbh *sqlx.DB
}

func NewHeatmapDB(dbh *sqlx.DB) HeatmapDB { return &heatmapDB{dbh: dbh} }

func (h *heatmapDB) GetAllHeatmapPointsInBbox(ctx context.Context, n float64, w float64, s float64, e float64) ([]model.HeatmapPoint, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("heatmap fetch data from bbox")

	query := `select dbm, st_asewkb(coordinates) as coordinates from "GsmHistory" 
				inner join public."GpsData" GD on GD.id = "GsmHistory".gps
				where st_x(coordinates) >= :N
				and st_x(coordinates) <= :W
				and st_y(coordinates) >= :S
				and st_y(coordinates) <= :E;`

	var heatmapPoints []model.HeatmapPoint

	if err := dbutils.NamedSelect(ctx, h.dbh, &heatmapPoints, query, map[string]interface{}{
		"N": n,
		"W": w,
		"S": s,
		"E": e,
	}); err != nil {
		return nil, err
	}

	return heatmapPoints, nil
}

func (h *heatmapDB) GetHeatmapPointsByIdDB(ctx context.Context, id int) (heatmapPoints []model.HeatmapPoint, err error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("heatmap fetch data from bbox")

	query := `select GH.dbm, st_asewkb(GPS.coordinates) as coordinates
				from "BaseStations"
    			inner join public."BsInfo" BA on "BaseStations".id = BA.bs
    			inner join public."arfcn" on BA.arfcn = arfcn.id
    			inner join public."GsmData" GD on arfcn.id = GD.arfcn
    			inner join public."GsmHistory" GH on GH.gsm = GD.id
    			inner join public."GpsData" GPS on GPS.id = GH.gps
    			where "BaseStations".id = :Id;`

	var heatmapPointsById []model.HeatmapPoint

	if err := dbutils.NamedSelect(ctx, h.dbh, &heatmapPointsById, query, map[string]interface{}{"Id": id}); err != nil {
		return nil, err
	}

	return heatmapPointsById, nil
}

func (h *heatmapDB) GetAllHeatmapPointsByCoordsDB(ctx context.Context, lat float64, lng float64) (heatmapPoints []model.HeatmapPoint, err error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("heatmap fetch data from bbox")

	query := `with Bs as (
    select id
    from (select id,
                 st_distance(coordinates, st_setsrid(st_makepoint(:Lng, :Lat), 4326)) distance
          from "BaseStations"
          order by distance
          limit 1) as inner_query
	)
	select GH.dbm, st_asewkb(GPS.coordinates) as coordinates 
	from Bs inner join "BsInfo" on Bs.id = "BsInfo".bs
        inner join "arfcn" on "BsInfo".arfcn = arfcn.id
        inner join "GsmData" on arfcn.id = "GsmData".arfcn
        inner join public."GsmHistory" GH on GH.gsm = "GsmData".id
        inner join public."GpsData" GPS on GPS.id = GH.gps;`

	var heatmapPointsById []model.HeatmapPoint

	if err := dbutils.NamedSelect(ctx, h.dbh, &heatmapPointsById, query, map[string]interface{}{"Lng": lng, "Lat": lat}); err != nil {
		return nil, err
	}

	return heatmapPointsById, nil
}
