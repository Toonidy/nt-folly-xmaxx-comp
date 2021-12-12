package db

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
)

// GetConnectionString grabs the db connection string from Viper Config settings
func GetConnectionString() string {
	connString := viper.GetString("db_url")
	if connString == "" {
		connString = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
			viper.GetString("db_user"),
			url.QueryEscape(viper.GetString("db_pass")),
			viper.GetString("db_host"),
			viper.GetString("db_port"),
			viper.GetString("db_name"),
			viper.GetString("db_sslmode"),
		)
	}
	return connString
}

// ConnectPool connects to a PostgreSQL Database using pgxpool.
func ConnectPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("could not parse db config: %w", err)
	}
	conn, err := pgxpool.ConnectConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("could not connect to db config: %w", err)
	}
	return conn, nil
}
