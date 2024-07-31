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
	GetPostPath(ctx context.Context, post *model.Post, date time.Time, measure int) error
	GetPostById(ctx context.Context, postId uuid.UUID) (*model.Post, error)
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

func (p *postDB) GetPostById(ctx context.Context, postId uuid.UUID) (*model.Post, error) {
	logger := logging.FromContext(ctx)
	logger.Debugw("post get by id")
	query := `select * from "Post" where id = :Id limit 1`
	var posts []model.Post
	if err := dbutils.NamedSelect(ctx, p.dbh, &posts, query, map[string]interface{}{"Id": postId.String()}); err != nil {
		return nil, err
	}
	if len(posts) != 0 {
		return &posts[0], nil
	}
	return nil, nil
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

func (p *postDB) GetPostPath(ctx context.Context, post *model.Post, date time.Time, measure int) error {
	logger := logging.FromContext(ctx)
	logger.Debugw("post scan times get")
	query := `select id, st_asewkb(coordinates) as coordinates, time, altitude, speed, heading, post_id from "GpsData" where post_id = :postId and cast(time as date) = :date order by time`
	var gpsData []model.GpsData

	if err := dbutils.NamedSelect(ctx, p.dbh, &gpsData, query, map[string]interface{}{"postId": post.Id, "date": date}); err != nil {
		return err
	}

	post.Coordinates = make([]model.GpsData, 0)
	if len(gpsData) != 0 {
		prev := gpsData[0]
		iterator := 0
		if iterator == measure {
			post.Coordinates = append(post.Coordinates, prev)
		}
		for i := 1; i < len(gpsData); i++ {
			curr := gpsData[i]
			timeDiff := curr.Time.Sub(prev.Time)
			if timeDiff.Seconds() > 120 {
				iterator++
				if iterator > measure {
					break
				}
			}
			if iterator == measure {
				post.Coordinates = append(post.Coordinates, curr)
			}
			prev = curr
		}
	}

	return nil
}
