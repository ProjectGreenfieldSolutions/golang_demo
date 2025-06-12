package db

import (
	"context"
	"fmt"
	"github.com/ProjectGreenfieldSolutions/golang_demo/models"
	"strings"
	"time"
)

// Test to see if the database has recently refreshed the data
func HasRecentFlightData(threshold time.Duration) (bool, error) {
	query := `SELECT created_at FROM flights ORDER BY created_at DESC LIMIT 1`

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var last time.Time
	err := DB.QueryRow(ctx, query).Scan(&last)
	if err != nil {
		return false, nil // No data means we should fetch
	}

	return time.Since(last) < threshold, nil
}

func SaveFlightsBatch(flights []models.Flight) error {
	if len(flights) == 0 {
		return nil
	}

	var (
		valueStrings []string
		valueArgs    []interface{}
	)

	for i, f := range flights {
		if f.Latitude == 0 && f.Longitude == 0 {
			continue
		}
		// (idx*8)+1 to offset placeholders per record
		valueStrings = append(valueStrings, fmt.Sprintf("($%d,$%d,$%d,to_timestamp($%d),$%d,$%d,$%d,$%d)",
			i*8+1, i*8+2, i*8+3, i*8+4, i*8+5, i*8+6, i*8+7, i*8+8))

		valueArgs = append(valueArgs,
			f.ICAO24,
			f.Callsign,
			f.OriginCountry,
			float64(f.TimePosition.Unix()),
			f.Latitude,
			f.Longitude,
			f.Altitude,
			f.Heading,
		)
	}

	if len(valueArgs) == 0 {
		return nil
	}

	query := fmt.Sprintf(`
		INSERT INTO flights (
			icao24, callsign, origin_country, time_position,
			lat, lng, altitude, heading
		) VALUES %s`, strings.Join(valueStrings, ","))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := DB.Exec(ctx, query, valueArgs...)
	if err != nil {
		fmt.Printf("❌ Batch insert failed: %v\n", err)
		return err
	}

	fmt.Printf("✅ Saved %d flights to DB\n", len(flights))
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
