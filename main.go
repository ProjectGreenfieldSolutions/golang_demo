package main

import (
	"context"
	"github.com/ProjectGreenfieldSolutions/golang_demo/db"
	"github.com/ProjectGreenfieldSolutions/golang_demo/models"
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"github.com/joho/godotenv"
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

	// Home route
	r.GET("/", func(c *gin.Context) {
		flights, err := FetchFlights()
		if err != nil {
			fmt.Printf("❌ Error fetching flights: %v", err)
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"title":   "Flight Tracker",
				"message": "Failed to load flights",
				"flights": nil,
			})
			return
		}

		// TODO - Batch insert version here
		for _, flight := range flights {
			_ = db.SaveFlightToDB(convertToStored(flight))
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Flight Tracker",
			"message": "Live flights from OpenSky API",
			"flights": flights,
		})
	})

	// API endpoint to return flights from DB
	r.GET("/flights", func(c *gin.Context) {
		rows, err := db.DB.Query(context.Background(), `
			SELECT id, icao24, callsign, origin_country, time_position,
			       lat, lng, altitude, heading, created_at
			FROM flights
			ORDER BY created_at DESC
			LIMIT 50
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
			return
		}
		defer rows.Close()

		var results []models.Flight
		for rows.Next() {
			var f models.Flight
			err := rows.Scan(
				&f.ID, &f.ICAO24, &f.Callsign, &f.OriginCountry,
				&f.TimePosition, &f.Latitude, &f.Longitude,
				&f.Altitude, &f.Heading, &f.CreatedAt,
			)
			if err == nil {
				results = append(results, f)
			}
		}
		c.JSON(http.StatusOK, results)
	})

	r.Run(":80")
}
