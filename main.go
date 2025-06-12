package main

import (
	"github.com/ProjectGreenfieldSolutions/golang_demo/db"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func main() {
	db.InitDB()
	defer db.CloseDB()

	if err := db.TestDBConnection(); err != nil {
		log.Fatal("🚨 Cannot proceed: DB is unreachable")
	}
	log.Printf("Reached DB")

	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	// Home route
	r.GET("/", func(c *gin.Context) {
		flights, err := FetchFlights()
		if err != nil {
			log.Printf("❌ Error fetching flights: %v", err)
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

	r.Run(":80")
}
