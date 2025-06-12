package models

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
	ICAO24        string
	Callsign      string
	OriginCountry string
	TimePosition  int64
	Latitude      float64
	Longitude     float64
	Altitude      float64
	Heading       float64
}
