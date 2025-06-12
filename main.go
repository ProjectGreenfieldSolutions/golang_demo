package main

import (
	"context"
	"github.com/ProjectGreenfieldSolutions/golang_demo/db"
	"github.com/ProjectGreenfieldSolutions/golang_demo/models"
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"github.com/joho/godotenv"
	"strconv"
	"time"
)

func main() {
	// Collect env variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("❌ Error loading .env file")
	}

	db.InitDB()

	if err := db.Migrate(); err != nil {
		fmt.Println("🚨 Migration failed, cannot start app")
	}
	
	defer db.CloseDB()

	if err := db.TestDBConnection(); err != nil {
		fmt.Println("🚨 Cannot proceed: DB is unreachable")
		return
	}
	fmt.Println("✅ Reached DB")

	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	startFlightFetcher(5 * 60 * time.Second)

	// Home route
	r.GET("/", func(c *gin.Context) {
		since := time.Now().Add(-1 * time.Hour)
	
		query := `
		SELECT DISTINCT ON (callsign) id, icao24, callsign, origin_country, time_position,
			   lat, lng, altitude, heading, created_at
		FROM flights
		WHERE time_position >= $1
		ORDER BY callsign, time_position DESC
		`
	
		rows, err := db.DB.Query(context.Background(), query, since)
		if err != nil {
			fmt.Printf("❌ Error querying flights: %v\n", err)
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"title":   "Flight Tracker",
				"message": "Failed to load stored flights",
				"flights": nil,
			})
			return
		}
		defer rows.Close()
	
		var flights []models.Flight
		for rows.Next() {
			var f models.Flight
			err := rows.Scan(
				&f.ID, &f.ICAO24, &f.Callsign, &f.OriginCountry,
				&f.TimePosition, &f.Latitude, &f.Longitude,
				&f.Altitude, &f.Heading, &f.CreatedAt,
			)
			if err == nil {
				flights = append(flights, f)
			}
		}
	
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Flight Tracker",
			"message": "Live flights (past 60 minutes)",
			"flights": flights,
		})
	})

	r.Run(":80")
}

func startFlightFetcher(interval time.Duration) {
	go func() {
		for {
			if shouldSkipFetch(interval) {
				fmt.Println("⏳ Recent data found — skipping API fetch.")
			} else {
				fetchAndStoreFlights()
			}
			time.Sleep(interval)
		}
	}()
}

func shouldSkipFetch(maxAge time.Duration) bool {
	var recent time.Time
	err := db.DB.QueryRow(context.Background(), "SELECT MAX(created_at) FROM flights").Scan(&recent)
	if err != nil {
		fmt.Printf("⚠️ Could not check for recent data: %v\n", err)
		return false
	}
	return time.Since(recent) < maxAge
}

func fetchAndStoreFlights() {
	flights, err := FetchFlights()
	if err != nil {
		fmt.Printf("❌ Fetch failed: %v\n", err)
		return
	}

	var stored []models.Flight
	for _, f := range flights {
		stored = append(stored, convertToStored(f))
	}
	if err := db.SaveFlightsBatch(stored); err != nil {
		fmt.Printf("❌ Batch insert failed: %v\n", err)
		return
	}
	fmt.Printf("📦 Stored %d flights\n", len(stored))
}
