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
		f.TimePosition,
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
