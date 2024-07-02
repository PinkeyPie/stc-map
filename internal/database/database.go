package database

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"simpleServer/internal/config"
)

type pgxLogger struct{}

func (pl pgxLogger) Log(ctx context.Context, level pgx.LogLevel, msg string, data map[string]interface{}) {
	var buffer bytes.Buffer
	buffer.WriteString(msg)

	for k, v := range data {
		buffer.WriteString(fmt.Sprintf("%s=%+v", k, v))
	}

	fmt.Println(buffer.String())
}

func NewDatabase(cfg *config.Config) (*sqlx.DB, error) {
	conn := flag.String("conn", cfg.DbConfig.DataSourceName, "database connection string")

	connConfig, err := pgx.ParseConfig(*conn)
	if err != nil {
		return nil, err
	}
	connConfig.RuntimeParams["application_name"] = "stc-map-api"

	connConfig.Logger = &pgxLogger{}
	connConfig.LogLevel = pgx.LogLevelDebug
	connStr := stdlib.RegisterConnConfig(connConfig)

	dbh, err := sqlx.Connect("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("prepare db connection: %w", err)
	}

	return dbh, nil
}
