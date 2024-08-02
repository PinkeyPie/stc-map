package database

import (
	"context"
	"encoding/json"
	"fmt"
	cluster "github.com/aliakseiz/gocluster"
	"github.com/jmoiron/sqlx"
	"log"
	"simpleServer/dbutils"
	"simpleServer/internal/baseStation/model"
	"simpleServer/internal/cache"
	"simpleServer/pkg/logging"
)

type BaseStationDB interface {
	Add(ctx context.Context, baseStation *model.BaseStation) error

	Update(ctx context.Context, id uint64, baseStation *model.BaseStation) error

	GetBaseStationById(ctx context.Context, id uint64) (*model.BaseStation, error)

	GetClusters(ctx context.Context, n float64, w float64, s float64, e float64, zoom float32) ([]cluster.Point, error)

	GetBaseStationByCoords(ctx context.Context, lat float64, lng float64) (*model.BaseStation, error)

	Fetch(ctx context.Context) ([]model.BaseStation, error)
}

type baseStationDB struct {
	dbh             *sqlx.DB
	cacheProvider   cache.ICacheProvider
	clusterProvider *cluster.Cluster
}

type latLng struct {
	Id  int64   `json:"id"`
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

func (tp latLng) GetCoordinates() *cluster.GeoCoordinates {
	return &cluster.GeoCoordinates{
		Lng: tp.Lng,
		Lat: tp.Lat,
	}
}

func (tp latLng) GetID() int64 {
	return tp.Id
}

type boundingBox struct {
	NW latLng `json:"nw"`
	SE latLng `json:"se"`
}

type ZoomInfo struct {
	Zoom int       `json:"zoom"`
	NW   []float64 `json:"nw"`
	SE   []float64 `json:"se"`
}

func createCluster(baseStations []model.BaseStation) *cluster.Cluster {
	fmt.Println("generating clusters")

	coords := make([]cluster.GeoPoint, len(baseStations))
	for i := range baseStations {
		lng, lat := baseStations[i].Coordinates.Point.X(), baseStations[i].Coordinates.Point.Y()
		coords[i] = latLng{
			Id:  int64(baseStations[i].ID),
			Lat: lat,
			Lng: lng,
		}
	}

	c, err := cluster.New(coords, cluster.WithinZoom(0, 15))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Clustering done")

	return c
}

func NewBaseStationDB(dbh *sqlx.DB, cacheProvider cache.ICacheProvider) BaseStationDB {
	ctx := context.Background()
	dbObtain := &baseStationDB{
		dbh:           dbh,
		cacheProvider: cacheProvider,
	}
	baseStations, err := dbObtain.Fetch(ctx)
	if err == nil {
		dbObtain.clusterProvider = createCluster(baseStations)
	}
	return dbObtain
}

func (bs *baseStationDB) Add(ctx context.Context, station *model.BaseStation) error {
	return nil
}

func (bs *baseStationDB) Update(ctx context.Context, id uint64, station *model.BaseStation) error {
	err := dbutils.RunTx(ctx, bs.dbh, func(tx *sqlx.Tx) error {
		//query := `select * from "BaseStations" where id = $1`
		//var baseStation model.BaseStation
		//if err := dbutils.Get(ctx, tx, &baseStation, query, station.ID); err != nil {
		//	return err
		//}
		//query = `update "BaseStations" set coordinates = st_makepoint($1, $2) where id = $3 returning *`
		//if err := dbutils.Get(ctx, tx, &baseStation, query, coords[0], coords[1], baseStation.ID); err != nil {
		//	return err
		//}

		return nil
	})

	return err
}

func (bs *baseStationDB) GetBaseStationById(ctx context.Context, id uint64) (*model.BaseStation, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("get base station by id", id)
	query := `select address, st_asewkb(coordinates) as coordinates, region, comment, id from "BaseStations" where id = :Id limit 1`
	var baseStations []model.BaseStation
	if err := dbutils.NamedSelect(ctx, bs.dbh, &baseStations, query, map[string]interface{}{"Id": id}); err != nil {
		return nil, err
	}
	if len(baseStations) != 0 {
		query = `select "BsInfo".* from "BsInfo" where "BsInfo".bs = :Bs`
		var bsInfo []model.BsInfo
		if err := dbutils.NamedSelect(ctx, bs.dbh, &bsInfo, query, map[string]interface{}{"Bs": id}); err == nil {
			baseStations[0].BsInfo = bsInfo
		}
		query = `select "Operators".* from "Operators" inner join "BsInfo" on "Operators".id = "BsInfo".operator_id where "BsInfo".bs = :Bs`
		var operators []model.Operator
		if err := dbutils.NamedSelect(ctx, bs.dbh, &operators, query, map[string]interface{}{"Bs": id}); err == nil {
			baseStations[0].Operators = operators
		}
		query = `select arfcn.id, arfcn_number, uplink, downlink, bandwidth, band, modulation, "CellularNetworkType".type as "CellularNetworkType"
				 from arfcn inner join "BsInfo" on arfcn.id = "BsInfo".arfcn 
				 inner join "CellularNetworkType" on arfcn."CellularNetworkType" = "CellularNetworkType".id
				 where bs = :Bs`
		var arfcns []model.Arfcn
		if err := dbutils.NamedSelect(ctx, bs.dbh, &arfcns, query, map[string]interface{}{"Bs": id}); err == nil {
			baseStations[0].Arfcn = arfcns
		}
		query = `select * from "Region" where id = :RegionId limit 1`
		var regions []model.Region
		if err := dbutils.NamedSelect(ctx, bs.dbh, &regions, query, map[string]interface{}{"RegionId": baseStations[0].RegionId}); err == nil {
			if len(regions) != 0 {
				baseStations[0].Region = regions[0]
			}
		}

		return &baseStations[0], nil
	}

	return nil, nil
}

func (bs *baseStationDB) GetBaseStationByCoords(ctx context.Context, lat, lng float64) (*model.BaseStation, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("Get base station by Coordinates", lat)
	query := `select address, coordinates, region, comment, id
			  	from (
					select address,
             			st_asewkb(coordinates) as                                                                       coordinates,
             			region,
             			comment,
             			id,
             			st_distance(coordinates, st_setsrid(st_makepoint(:Lng, :Lat), 4326)) distance
      				from "BaseStations"
      				order by distance
      				limit 1)
				as inner_query;`
	var baseStation []model.BaseStation
	if err := dbutils.NamedSelect(ctx, bs.dbh, &baseStation, query, map[string]interface{}{"Lat": lat, "Lng": lng}); err != nil {
		return nil, err
	}
	if len(baseStation) != 0 {
		query = `select "BsInfo".* from "BsInfo" where "BsInfo".bs = :Bs`
		var bsInfo []model.BsInfo
		if err := dbutils.NamedSelect(ctx, bs.dbh, &bsInfo, query, map[string]interface{}{"Bs": baseStation[0].ID}); err == nil {
			baseStation[0].BsInfo = bsInfo
		}
		query = `select "Operators".* from "Operators" inner join "BsInfo" on "Operators".id = "BsInfo".operator_id where "BsInfo".bs = :Bs`
		var operators []model.Operator
		if err := dbutils.NamedSelect(ctx, bs.dbh, &operators, query, map[string]interface{}{"Bs": baseStation[0].ID}); err == nil {
			baseStation[0].Operators = operators
		}
		query = `select arfcn.id, arfcn_number, uplink, downlink, bandwidth, band, modulation, "CellularNetworkType".type as "CellularNetworkType"
				 from arfcn inner join "BsInfo" on arfcn.id = "BsInfo".arfcn 
				 inner join "CellularNetworkType" on arfcn."CellularNetworkType" = "CellularNetworkType".id
				 where bs = :Bs`
		var arfcns []model.Arfcn
		if err := dbutils.NamedSelect(ctx, bs.dbh, &arfcns, query, map[string]interface{}{"Bs": baseStation[0].ID}); err == nil {
			baseStation[0].Arfcn = arfcns
		}
		query = `select * from "Region" where id = :RegionId limit 1`
		var regions []model.Region
		if err := dbutils.NamedSelect(ctx, bs.dbh, &regions, query, map[string]interface{}{"RegionId": baseStation[0].RegionId}); err == nil {
			if len(regions) != 0 {
				baseStation[0].Region = regions[0]
			}
		}

		return &baseStation[0], nil
	}

	return nil, nil
}

func (bs *baseStationDB) Fetch(ctx context.Context) ([]model.BaseStation, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("base stations fetch all")
	query := `select id, address, st_asewkb(coordinates) as coordinates, region, comment from "BaseStations"`

	var baseStations []model.BaseStation
	if err := dbutils.Select(ctx, bs.dbh, &baseStations, query); err != nil {
		return nil, err
	}

	return baseStations, nil
}

func (bs *baseStationDB) GetClusters(ctx context.Context, n float64, w float64, s float64, e float64, zoom float32) ([]cluster.Point, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("get clusters")
	origin := "origin" // need to find client value for redis

	zoomInfo, err := bs.zoomInCache(ctx, origin)
	if err == nil && zoomInfo != nil {
		if int(zoom) == zoomInfo.Zoom && n <= zoomInfo.NW[0] && w >= zoomInfo.NW[1] && s >= zoomInfo.SE[0] && e <= zoomInfo.SE[1] {
			points, err := bs.clusterInCache(ctx, origin, zoomInfo)
			if err == nil && len(points) != 0 {
				return points, nil
			}
		}
	}

	var points []cluster.Point
	if zoom >= 15 {
		query := `select id, st_x(coordinates) as X, st_y(coordinates) as Y, 1 as NumPoints from "BaseStations" where st_x(coordinates) > :w and st_x(coordinates) < :e and st_y(coordinates) > :s and st_y(coordinates) < :n`
		if err = dbutils.NamedSelect(ctx, bs.dbh, &points, query, map[string]interface{}{"n": n, "w": w, "s": s, "e": e}); err != nil {
			return nil, err
		}
	} else {
		nw := latLng{Lat: n, Lng: w}
		se := latLng{Lat: s, Lng: e}
		points, _ = bs.clusterProvider.GetClusters(nw, se, int(zoom), -1)
	}
	zoomInfo = &ZoomInfo{
		Zoom: int(zoom),
		NW:   []float64{n, w},
		SE:   []float64{s, e},
	}
	err = bs.zoomSaveInCache(ctx, origin, zoomInfo)
	if err == nil {
		err = bs.clusterSaveInCache(ctx, points, origin, zoomInfo)
		if err != nil {
			logger.Debugw("Error while save in cache")
		}
	}
	return points, nil
}

func (bs *baseStationDB) zoomInCache(ctx context.Context, origin string) (*ZoomInfo, error) {
	if cache.IsCacheSkip(ctx) {
		return nil, nil
	}

	var (
		item string = ""
		key         = bs.zoomByOrigin(origin)
	)

	exists, _ := bs.cacheProvider.Exists(ctx, key)
	if exists {
		err := bs.cacheProvider.Fetch(ctx, key, &item, nil)
		if err != nil {
			return nil, err
		}
	}
	if item != "" {
		var zoom ZoomInfo
		err := json.Unmarshal([]byte(item), &zoom)
		if err == nil {
			return &zoom, nil
		}
	}

	return nil, nil
}

func (bs *baseStationDB) zoomSaveInCache(ctx context.Context, origin string, zoom *ZoomInfo) error {
	key := bs.zoomByOrigin(origin)
	js, err := json.Marshal(zoom)
	if err == nil {
		err = bs.cacheProvider.Set(ctx, key, js)
	}
	if err != nil {
		return err
	}
	return nil
}

func (bs *baseStationDB) clusterInCache(ctx context.Context, origin string, zoom *ZoomInfo) ([]cluster.Point, error) {
	if cache.IsCacheSkip(ctx) {
		return nil, nil
	}
	var (
		item string = ""
		key         = bs.clusterByCoords(origin, zoom.NW, zoom.SE, zoom.Zoom)
	)
	if exists, _ := bs.cacheProvider.Exists(ctx, key); exists {
		err := bs.cacheProvider.Fetch(ctx, key, &item, nil)
		if err != nil {
			return nil, err
		}
		var points []cluster.Point
		err = json.Unmarshal([]byte(item), &points)
		if err == nil {
			return points, nil
		}
	}

	return nil, nil
}

func (bs *baseStationDB) clusterSaveInCache(ctx context.Context, points []cluster.Point, origin string, zoom *ZoomInfo) error {
	key := bs.clusterByCoords(origin, zoom.NW, zoom.SE, zoom.Zoom)
	js, err := json.Marshal(points)
	if err == nil {
		err = bs.cacheProvider.Set(ctx, key, js)
	}
	if err != nil {
		return err
	}
	return nil
}

const (
	cacheKeyBsById          = "bs-by-id"
	cacheKeyClusterByCoords = "cluster-by-coords"
	cacheKeyZoomByOrigin    = "zoom-by-origin"
)

func (bs *baseStationDB) bsByIdCacheKey(id uint64) string {
	return fmt.Sprintf("%s.%d", cacheKeyBsById, id)
}

func (bs *baseStationDB) clusterByCoords(origin string, nw []float64, se []float64, zoom int) string {
	return fmt.Sprintf("%s.%s.%f.%f.%f.%f.%d", cacheKeyClusterByCoords, origin, nw[0], nw[1], se[0], se[1], zoom)
}

func (bs *baseStationDB) zoomByOrigin(origin string) string {
	return fmt.Sprintf("%s.%s", cacheKeyZoomByOrigin, origin)
}

func exampleSimpleSelect(ctx context.Context, dbh *sqlx.DB) error {
	var baseStations []model.BaseStation
	query := `select * from "BaseStations"`
	if err := dbutils.Select(ctx, dbh, &baseStations, query); err != nil {
		return err
	}
	log.Println(baseStations)

	return nil
}

func exampleSelectWith(ctx context.Context, dbh *sqlx.DB, coords []float64) error {
	if len(coords) != 2 {
		return fmt.Errorf("Error in passing coordinates")
	}
	var baseStations []model.BaseStation
	query := `select * from "BaseStations" where st_x(coordinates) = $1 and st_y(coordinates) = $2`
	if err := dbutils.Select(ctx, dbh, &baseStations, query, coords[0], coords[1]); err != nil {
		return err
	}
	log.Println(baseStations)

	return nil
}
