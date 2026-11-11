package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vkmnamit/cloudx/configs"
)

func Connect() (*pgxpool.Pool, error) {

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		configs.AppConfig.DBUser,
		configs.AppConfig.DBPassword,
		configs.AppConfig.DBHost,
		configs.AppConfig.DBPort,
		configs.AppConfig.DBName,
		configs.AppConfig.DBSSLMode,
	)

	conn, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		conn.Close()
		return nil, err
	}

	fmt.Println("✅ Connected to PostgreSQL")

	return conn, nil
}
