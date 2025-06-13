package main

import (
	"context"
	"github.com/ProjectGreenfieldSolutions/golang_demo/db"
	"github.com/ProjectGreenfieldSolutions/golang_demo/models"
	"github.com/jackc/pgx/v5"
	"github.com/gin-gonic/gin"
	"fmt"
	"net/http"
	"github.com/joho/godotenv"
	"time"
	"strings"
	"os"
)

func main() {
	// Collect env variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("❌ Error loading .env file")
	}

	mode := os.Getenv("APP_MODE")

	if mode == "" {
		mode = gin.ReleaseMode // fallback if not set
	} 

	gin.SetMode(mode)

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
		page := strings.ToUpper(c.Query("page"))
		letter := ""
		since := time.Now().Add(-1 * time.Hour)
		tabletitlemessage := ""
		alphabet := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		letters := make([]int, len(alphabet))

		// Move these out to "global" of this function
		var (
			rows      pgx.Rows
			query_err  error
			flights []models.Flight
		)

		if page != "" {
			letter = page + "%"
			query := `
			SELECT DISTINCT ON (callsign) id, icao24, callsign, origin_country,
				time_position, lat, lng, altitude, heading, created_at
			FROM flights
			WHERE time_position >= $1
		  		AND callsign ILIKE $2
			ORDER BY callsign, time_position DESC
			`
			rows, query_err = db.DB.Query(context.Background(), query, since, letter)
		} else {
			query := `
			SELECT DISTINCT ON (callsign) id, icao24, callsign, origin_country,
				time_position, lat, lng, altitude, heading, created_at
			FROM flights
			WHERE time_position >= $1
			ORDER BY callsign, time_position DESC
			`
			rows, query_err = db.DB.Query(context.Background(), query, since)
		}

		if query_err != nil {
			fmt.Printf("❌ Error querying flights: %v\n", query_err)
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"title":   "Flight Tracker",
				"message": "Failed to load stored flights",
				"flights": nil,
			})
			return
		}

		defer rows.Close()
	
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

		for i := range letters {
			letters[i] = i
		}
		
		if page != "" {
			tabletitlemessage = fmt.Sprintf("Displaying %d flights that start with the letter %s", len(flights), page) 
		} else {
			tabletitlemessage = fmt.Sprintf("Displaying ALL %d flights", len(flights)) 
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":    "Flight Tracker",
			"message":  fmt.Sprintf("Flights around Detroit since (%s)", since.Format("Jan 2, 3:04PM")),
			"tabletitle": tabletitlemessage,
			"flights":  flights,
			"page":     string(page),
			"alphabet": alphabet,
			"letters":  letters,
		})
	})

	r.GET("/api/path/:callsign", func(c *gin.Context) {
		callsign := c.Param("callsign")
		fmt.Printf("📥 Received path request for callsign: %s\n", callsign)
	
		query := `
		SELECT callsign, lat, lng, altitude, heading, created_at
		FROM flights
		WHERE TRIM(callsign) = $1
		ORDER BY created_at ASC
		`
		fmt.Println("🔍 Executing path query...")
	
		rows, err := db.DB.Query(context.Background(), query, callsign)
		if err != nil {
			fmt.Printf("❌ DB query error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB query failed"})
			return
		}
		defer rows.Close()
	
		var path []models.Point
		for rows.Next() {
			var p models.Point
			if err := rows.Scan(&p.Callsign, &p.Lat, &p.Lng, &p.Altitude, &p.Heading, &p.Timestamp); err != nil {
				fmt.Printf("⚠️ Row scan error: %v\n", err)
				continue
			}
			fmt.Printf("➡️ Appending point: %+v\n", p)
			path = append(path, p)
		}
	
		fmt.Printf("✅ Total points collected: %d\n", len(path))
		c.JSON(http.StatusOK, path)
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
