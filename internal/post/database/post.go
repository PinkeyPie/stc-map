package database

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/jmoiron/sqlx"
	"simpleServer/dbutils"
	"simpleServer/internal/post/model"
	"simpleServer/pkg/logging"
	"time"
)

type PostDB interface {
	GetAllPosts(ctx context.Context) ([]model.Post, error)
	GetPostScanDates(ctx context.Context, postId uuid.UUID) ([]model.GpsData, error)
	GetPostPath(ctx context.Context, postId uuid.UUID, date time.Time) (*model.Post, error)
}

type postDB struct {
	dbh *sqlx.DB
}

func NewPostDB(dbh *sqlx.DB) PostDB {
	return &postDB{dbh: dbh}
}

func (p *postDB) GetAllPosts(ctx context.Context) ([]model.Post, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("post fetch all")
	query := `select * from "Post"`

	var posts []model.Post
	if err := dbutils.Select(ctx, p.dbh, &posts, query); err != nil {
		return nil, err
	}

	return posts, nil
}

func (p *postDB) GetPostScanDates(ctx context.Context, postId uuid.UUID) ([]model.GpsData, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("post dates get")
	query := `select id, st_asewkb(coordinates) as coordinates, time, altitude, speed, heading, post_id from "GpsData" where post_id = :PostId order by time`
	var gpsData []model.GpsData

	if err := dbutils.NamedSelect(ctx, p.dbh, &gpsData, query, map[string]interface{}{"PostId": postId}); err != nil {
		return nil, err
	}
	return gpsData, nil
}

func (p *postDB) GetPostPath(ctx context.Context, postId uuid.UUID, date time.Time) (*model.Post, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("post scan times get")

	return nil, nil
}
