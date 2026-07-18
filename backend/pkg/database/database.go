package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vkmnamit/cloudx/configs"
)

func Connect() (*pgx.Conn, error) {

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		configs.AppConfig.DBUser,
		configs.AppConfig.DBPassword,
		configs.AppConfig.DBHost,
		configs.AppConfig.DBPort,
		configs.AppConfig.DBName,
		configs.AppConfig.DBSSLMode,
	)

	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	fmt.Println("✅ Connected to PostgreSQL")

	return conn, nil
}
