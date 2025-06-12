package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

var DB *pgxpool.Pool

func InitDB() {
	dsn := os.Getenv("DATABASE_URL")
	var err error
	DB, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		fmt.Printf("❌ Failed to connect to DB: %v", err)
	}
	fmt.Println("✅ Connected to PostgreSQL")
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}
