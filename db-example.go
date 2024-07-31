package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	uuid "github.com/jackc/pgtype/ext/gofrs-uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/twpayne/go-geom/encoding/ewkb"
	"go.uber.org/multierr"
	"log"
	"simpleServer/dbutils"
	"time"
)

var conn = flag.String("conn", "postgres://postgres:stcspb@localhost/observer?sslmode=disable", "database connection string")

type pgxLogger struct{}

type Post struct {
	ID      uuid.UUID `db:"id"`
	Name    string    `db:"name"`
	GpsData []GpsData
}

type GpsData struct {
	ID          uuid.UUID  `db:"id"`
	Coordinates ewkb.Point `db:"coordinates"`
	Time        time.Time  `db:"time"`
	Altitude    float32    `db:"altitude"`
	Speed       float32    `db:"speed"`
	Heading     float32    `db:"heading"`
	PostId      uuid.UUID  `db:"post_id"`
}

func exampleCoords(ctx context.Context, dbh *sqlx.DB) error {
	query := `select * from "Post"`
	var posts []Post
	err := dbutils.Select(ctx, dbh, &posts, query)
	if err != nil {
		return err
	}

	query = `select "GpsData".id, st_asewkb(coordinates) as coordinates, time, altitude, speed, heading, post_id from "GpsData"`
	var coords []GpsData
	err = dbutils.Select(ctx, dbh, &coords, query)
	if err != nil {
		return err
	}
	log.Println(coords)

	return nil
}

func exampleJoin(ctx context.Context, dbh *sqlx.DB) error {
	query := `select * from "Post"`
	var posts []Post
	err := dbutils.Select(ctx, dbh, &posts, query)
	if err != nil {
		return err
	}
	for i := range posts {
		query = `select id, st_asewkb(coordinates) as coordinates, time, altitude, speed, heading, post_id 
				 from "GpsData" where post_id = $1`
		var coords []GpsData
		if err := dbutils.Select(ctx, dbh, &coords, query, posts[i].ID); err != nil {
			return err
		}
		posts[i].GpsData = coords
	}

	log.Println(posts)

	return nil
}

func (pl *pgxLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	var buffer bytes.Buffer
	buffer.WriteString(msg)

	for k, v := range data {
		buffer.WriteString(fmt.Sprintf(" %s=%+v", k, v))
	}

	log.Println(buffer.String())
}

func run() error {
	ctx := context.Background()

	connConfig, err := pgx.ParseConfig(*conn)
	if err != nil {
		return err
	}
	connConfig.RuntimeParams["application_name"] = "db-example"

	connConfig.Logger = &pgxLogger{}
	connConfig.LogLevel = pgx.LogLevelDebug
	connStr := stdlib.RegisterConnConfig(connConfig)

	dbh, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return fmt.Errorf("prepare db connection: %w", err)
	}
	defer dbh.Close()

	return exampleJoin(ctx, dbh)
}

type User struct {
	ID    int64  `db:"id"`
	Login string `db:"login"`
	Name  string `db:"name"`
}

var initScript = `
drop table if exists test_users;
create table test_users (
   id bigserial,
   login text,
   name text
);

insert into test_users (login, name) values
('ivanov', 'Иванов Иван Иванович'),
('petrov', 'Petrov Petr Sidorovich'),
('sidorov', 'Sidorov Sidor Sidorovich')`

func realExample(ctx context.Context, dbh *sqlx.DB) error {
	posts := make([]Post, 0)
	q := `select * from "Post"`
	rows, err := dbh.QueryxContext(ctx, q)
	if err != nil {
		return err
	}
	defer func() {
		err = multierr.Combine(err, rows.Close())
	}()

	for rows.Next() {
		var post Post
		if err := rows.StructScan(&post); err != nil {
			return err
		}

		posts = append(posts, post)
	}
	if rows.Err() != nil {
		return rows.Err()
	}

	posts = make([]Post, 0)
	q = `select * from "Post"`
	if err = dbh.SelectContext(ctx, &posts, q); err != nil {
		return err
	}

	posts = make([]Post, 0)
	q = `select * from "Post"`
	if err = dbh.SelectContext(ctx, &posts, q); err != nil {
		return err
	}

	postsMap := make(map[string]Post, len(posts))
	for _, post := range posts {
		postsMap[post.Name] = post
	}

	return nil
}

func example(ctx context.Context, dbh *sqlx.DB) error {
	if _, err := dbutils.Exec(ctx, dbh, initScript); err != nil {
		return err
	}

	var users []User
	q := `select * from test_users`
	if err := dbutils.Select(ctx, dbh, &users, q); err != nil {
		return err
	}
	log.Println(users)

	mm, err := dbutils.SelectMaps(ctx, dbh, q)
	if err != nil {
		return err
	}
	log.Println(mm)

	q = `select * from test_users where login=$1`
	m, err := dbutils.GetMap(ctx, dbh, q, "ivanov")
	if err != nil {
		return err
	}
	log.Println(m)

	q = `select * from test_users where login=:login`
	m, err = dbutils.NamedGetMap(ctx, dbh, q, map[string]interface{}{"Login": "petrov"})
	if err != nil {
		return err
	}
	log.Println(m)

	var users2 []User
	q = `select * from test_users`
	if err = dbh.SelectContext(ctx, &users2, q); err != nil {
		return err
	}

	q = `select * from test_users where login = any($1)`
	if err := dbutils.Select(ctx, dbh, &users, q, []string{"ivanov", "petrov"}); err != nil {
		return err
	}
	log.Println(users)

	u, err := updateUser(ctx, dbh, "ivanov", "Sergeev")
	if err != nil {
		return err
	}
	log.Println(u)

	return nil
}

func updateUser(ctx context.Context, dbh *sqlx.DB, login string, newName string) (u User, err error) {
	err = dbutils.RunTx(ctx, dbh, func(tx *sqlx.Tx) error {
		u, err = updateUserTx(ctx, tx, login, newName)
		return err
	})

	return u, err
}

func updateUserTx(ctx context.Context, tx sqlx.ExtContext, login string, newName string) (u User, err error) {
	q := `select * from test_users where login = $1`
	if err := dbutils.Get(ctx, tx, &u, q, login); err != nil {
		return User{}, err
	}

	q = `update test_users set name = $1 where login = $2 returning *`
	if err := dbutils.Get(ctx, tx, &u, q, newName, login); err != nil {
		return User{}, err
	}

	return u, nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("errpr: %+v", err)
	}
}
