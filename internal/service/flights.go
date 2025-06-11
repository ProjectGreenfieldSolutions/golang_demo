package service

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

    flights := make([]Flight, len(data.States))
    for i, s := range data.States {
        flights[i] = Flight{
            ICAO24:        fmt.Sprintf("%v", s[0]),
            Callsign:      fmt.Sprintf("%v", s[1]),
            OriginCountry: fmt.Sprintf("%v", s[2]),
            TimePosition:  toInt64(s[3]),
            LastContact:   toInt64(s[4]),
            Longitude:     toFloat64(s[5]),
            Latitude:      toFloat64(s[6]),
            Altitude:      toFloat64(s[7]),
            OnGround:      s[8].(bool),
            Velocity:      toFloat64(s[9]),
            Heading:       toFloat64(s[10]),
            VerticalRate:  toFloat64(s[11]),
        }
    }
    return flights, nil
}

// Helpers: Convert interface{} to numerical types
func toInt64(v interface{}) int64    { if v == nil { return 0 }; return int64(v.(float64)) }
func toFloat64(v interface{}) float64 { if v == nil { return 0 }; return v.(float64) }

