package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ProjectGreenfieldSolutions/golang_demo/models"
	"golang.org/x/oauth2/clientcredentials"
	"io"
	"net/http"
	"os"
	"time"
)

type StateVectorResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}

func FetchFlights() ([]models.OpenSkyFlight, error) {

	cfg := clientcredentials.Config{
		ClientID:     os.Getenv("OPEN_SKY_CLIENT_ID"),
		ClientSecret: os.Getenv("OPEN_SKY_CLIENT_SECRET"),
		TokenURL:     "https://auth.opensky-network.org/auth/realms/opensky-network/protocol/openid-connect/token",
	}

	ctx := context.Background()

	client := cfg.Client(ctx)

	resp, err := client.Get(os.Getenv("OPEN_SKY_COORDINATES"))

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("❌ OpenSky API failed — status %d, body: %s", resp.StatusCode, string(body))

		return []models.OpenSkyFlight{
			{
				ICAO24:        "no_data",
				Callsign:      "API_ERROR",
				OriginCountry: "Unavailable",
				TimePosition:  time.Now().Unix(),
				LastContact:   time.Now().Unix(),
				Longitude:     0.0,
				Latitude:      0.0,
				Altitude:      0.0,
				OnGround:      true,
				Velocity:      0.0,
				Heading:       0.0,
				VerticalRate:  0.0,
			},
		}, nil
	}

	var data StateVectorResponse

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	flights := make([]models.OpenSkyFlight, 0, len(data.States))
	for _, s := range data.States {
		flights = append(flights, models.OpenSkyFlight{
			ICAO24:        toStr(s[0]),
			Callsign:      toStr(s[1]),
			OriginCountry: toStr(s[2]),
			TimePosition:  toInt64(s[3]),
			LastContact:   toInt64(s[4]),
			Longitude:     toFloat64(s[5]),
			Latitude:      toFloat64(s[6]),
			Altitude:      toFloat64(s[7]),
			OnGround:      toBool(s[8]),
			Velocity:      toFloat64(s[9]),
			Heading:       toFloat64(s[10]),
			VerticalRate:  toFloat64(s[11]),
		})
	}
	fmt.Println("✅ Fetched flight data with OAuth2")
	return flights, nil
}

func convertToStored(f models.OpenSkyFlight) models.Flight {
	return models.Flight{
		ICAO24:        f.ICAO24,
		Callsign:      f.Callsign,
		OriginCountry: f.OriginCountry,
		TimePosition:  time.Unix(f.TimePosition, 0),
		Latitude:      f.Latitude,
		Longitude:     f.Longitude,
		Altitude:      f.Altitude,
		Heading:       f.Heading,
	}
}

func toStr(v interface{}) string {
	if v == nil {
		return ""
	}
	s := fmt.Sprintf("%v", v)
	// fmt.Printf("❌ toStr: %s", s)
	return s
}

func toInt64(v interface{}) int64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return int64(f)
	}
	// fmt.Printf("❌ toInt64: unexpected type %T", v)
	return 0
}

func toFloat64(v interface{}) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	// fmt.Printf("❌ toFloat64: unexpected type %T", v)
	return 0
}

func toBool(v interface{}) bool {
	if v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return b
	}
	// fmt.Printf("❌ toBool: unexpected type %T", v)
	return false
}
