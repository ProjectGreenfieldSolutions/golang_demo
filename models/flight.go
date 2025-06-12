package models

import "time"

type Point struct {
	Callsign      string  `json:"callsign"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	Heading   float64   `json:"heading"`
	Altitude  float64   `json:"altitude"`
	Timestamp time.Time `json:"timestamp"`
}

type OpenSkyFlight struct {
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

type Flight struct {
	ID            int       `json:"id"`
	ICAO24        string    `json:"icao24"`
	Callsign      string    `json:"callsign"`
	OriginCountry string    `json:"origin_country"`
	TimePosition  time.Time `json:"time_position"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Altitude      float64   `json:"altitude"`
	Heading       float64   `json:"heading"`
	CreatedAt     time.Time `json:"created_at"`
}
