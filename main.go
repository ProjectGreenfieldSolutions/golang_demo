package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	// Home route
	r.GET("/", func(c *gin.Context) {
		flights, err := FetchFlights()
		if err != nil {
			c.HTML(http.StatusInternalServerError, "index.html", gin.H{
				"title":   "Flight Tracker",
				"message": "Failed to load flights",
				"flights": nil,
			})
			return
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":   "Flight Tracker",
			"message": "Live flights from OpenSky API",
			"flights": flights[:20],
		})
	})

	r.Run(":8080")
}

