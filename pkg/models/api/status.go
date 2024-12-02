package api

import "time"

type Uptime struct {
	Since        time.Time `json:"since"`
	Milliseconds uint64    `json:"milliseconds"`
}

type Status struct {
	Version string    `json:"version"`
	Runtime string    `json:"runtime"`
	Mode    string    `json:"mode"`
	Time    time.Time `json:"time"`
	Uptime  Uptime    `json:"uptime"`
}
