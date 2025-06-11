package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const OpenSkyURL = "https://opensky-network.org/api/states/all"

type StateVectorResponse struct {
	Time   int64           `json:"time"`
	States [][]interface{} `json:"states"`
}

type Flight struct {
	ICAO24        string  `json:"icao24"`
	Callsign      string  `json:"callsign"`
	OriginCountry string  `json:"origin_country"`
	TimePosition  int64   `json:"time_position"`
	LastContact   int64   `json:"last_contact"`
	Longitude     float64 `json:"longitude"`
	Latitude      float64 `json:"latitude"`
	Altitude      float64 `json:"altitude"`
	OnGround      bool    `json:"on_ground"`
	Velocity      float64 `json:"velocity"`
	Heading       float64 `json:"heading"`
	VerticalRate  float64 `json:"vertical_rate"`
}

func FetchFlights() ([]Flight, error) {
	resp, err := http.Get(OpenSkyURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data StateVectorResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	flights := make([]Flight, 0, len(data.States))
	for _, s := range data.States {
		flights = append(flights, Flight{
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
	return flights, nil
}

// Helpers
func toStr(v interface{}) string     { if v == nil { return "" }; return fmt.Sprintf("%v", v) }
func toInt64(v interface{}) int64    { if v == nil { return 0 }; return int64(v.(float64)) }
func toFloat64(v interface{}) float64 { if v == nil { return 0 }; return v.(float64) }
func toBool(v interface{}) bool      { if v == nil { return false }; return v.(bool) }

