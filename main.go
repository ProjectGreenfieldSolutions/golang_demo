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

	startFlightFetcher(30 * time.Second)

	// Home route
	r.GET("/", func(c *gin.Context) {
		page := strings.ToUpper(c.Query("page"))
		since := time.Hour
	
		flights, err := FetchRecentFlights(since, page)
		if err != nil {
			fmt.Printf("❌ Error querying flights: %v\n", err)
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"title":   "Flight Tracker",
				"message": "Failed to load stored flights",
				"flights": nil,
			})
			return
		}
	
		tabletitle := fmt.Sprintf("Displaying ALL %d flights", len(flights))
		if page != "" {
			tabletitle = fmt.Sprintf("Displaying %d flights that start with the letter %s", len(flights), page)
		}
	
		alphabet := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		letters := make([]int, len(alphabet))
		for i := range letters {
			letters[i] = i
		}
	
		// Get the locale time
		loc, _ := time.LoadLocation("America/Detroit")
		formattedTime := time.Now().In(loc).Format("Jan 2, 3:04PM MST")

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":      "Flight Tracker",
			"message": 	  fmt.Sprintf("Flights around Detroit since (%s)", formattedTime),
			"tabletitle": tabletitle,
			"flights":    flights,
			"page":       page,
			"alphabet":   alphabet,
			"letters":    letters,
		})
	})

	r.GET("/api/flights", func(c *gin.Context) {
		page := strings.ToUpper(c.Query("page"))
		flights, err := FetchRecentFlights(1*time.Hour, page)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve flights"})
			return
		}
		c.JSON(http.StatusOK, flights)
	})

	r.GET("/api/path/:callsign", func(c *gin.Context) {
		callsign := c.Param("callsign")
		fmt.Printf("📥 Received path request for callsign: %s\n", callsign)
	
		query := `
		SELECT callsign, lat, lng, altitude, heading, created_at
		FROM flights
		WHERE TRIM(callsign) = $1
			AND created_at >= $2
		ORDER BY created_at ASC
		`
		fmt.Println("🔍 Executing path query...")

		since := time.Hour
		timePositionAfter := time.Now().Add(-since)
	
		rows, err := db.DB.Query(context.Background(), query, callsign, timePositionAfter)
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
				deleteOldFlights()
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

func deleteOldFlights() error {
	query := `
		DELETE FROM flights
		WHERE created_at < NOW() - INTERVAL '1 hours'
	`
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := db.DB.Exec(ctx, query)
	if err != nil {
		fmt.Printf("❌ Failed to delete old flights: %v\n", err)
		return err
	}

	rowsDeleted := res.RowsAffected()
	fmt.Printf("Deleted %d old flight records\n", rowsDeleted)
	return nil
}

func FetchRecentFlights(since time.Duration, page string) ([]models.Flight, error) {
	var (
		rows pgx.Rows
		err  error
	)

	createdAfter := time.Now().Add(-30 * time.Second)
	timePositionAfter := time.Now().Add(-since)

	if page != "" {
		letter := page + "%"
		query := `
		SELECT DISTINCT ON (callsign) id, icao24, callsign, origin_country,
			time_position, lat, lng, altitude, heading, created_at
		FROM flights
		WHERE time_position >= $1 AND callsign ILIKE $2 AND created_at >= $3
		ORDER BY callsign, time_position DESC`
		rows, err = db.DB.Query(context.Background(), query, timePositionAfter, letter, createdAfter)
	} else {
		query := `
		SELECT DISTINCT ON (callsign) id, icao24, callsign, origin_country,
			time_position, lat, lng, altitude, heading, created_at
		FROM flights
		WHERE time_position >= $1 AND created_at >= $2
		ORDER BY callsign, time_position DESC`
		rows, err = db.DB.Query(context.Background(), query, timePositionAfter, createdAfter)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flights []models.Flight
	for rows.Next() {
		var f models.Flight
		if scanErr := rows.Scan(
			&f.ID, &f.ICAO24, &f.Callsign, &f.OriginCountry,
			&f.TimePosition, &f.Latitude, &f.Longitude,
			&f.Altitude, &f.Heading, &f.CreatedAt,
		); scanErr == nil {
			flights = append(flights, f)
		}
	}

	return flights, nil
}
