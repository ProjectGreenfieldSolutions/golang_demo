package db

import (
	"context"
	"fmt"
	"github.com/ProjectGreenfieldSolutions/golang_demo/models"
	"time"
)

func SaveFlightToDB(f models.Flight) error {
	if f.Latitude == 0 && f.Longitude == 0 {
		return nil
	}

	query := `
		INSERT INTO flights (
			icao24, callsign, origin_country, time_position,
			lat, lng, altitude, heading
		) VALUES ($1, $2, $3, to_timestamp($4), $5, $6, $7, $8)
	`

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := DB.Exec(ctx, query,
		f.ICAO24,
		f.Callsign,
		f.OriginCountry,
		float64(f.TimePosition.Unix()),
		f.Latitude,
		f.Longitude,
		f.Altitude,
		f.Heading,
	)

	if err != nil {
		fmt.Printf("❌ DB insert failed for %s: %v\n", f.ICAO24, err)
		return err
	}

	fmt.Printf("📡 Saved flight %s to DB\n", f.ICAO24)
	return nil
}

func TestDBConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := DB.Ping(ctx)
	if err != nil {
		fmt.Println("❌ DB connection test failed:", err)
	} else {
		fmt.Println("✅ DB connection is alive")
	}
	return err
}

func Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS flights (
		id SERIAL PRIMARY KEY,
		icao24 TEXT,
		callsign TEXT,
		origin_country TEXT,
		time_position TIMESTAMP,
		lat DOUBLE PRECISION,
		lng DOUBLE PRECISION,
		altitude DOUBLE PRECISION,
		heading DOUBLE PRECISION,
		created_at TIMESTAMPTZ DEFAULT now()
	)`
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := DB.Exec(ctx, query)
	if err != nil {
		fmt.Printf("❌ Migration failed: %v", err)
		return err
	}

	fmt.Println("✅ Schema migrated (flights table ensured)")
	return nil
}
