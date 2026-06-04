package store

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"os"
)

func Connect() (*pgxpool.Pool, error) {
	return pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
}
