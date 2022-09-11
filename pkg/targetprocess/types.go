package targetprocess

import "time"

// RuntimeInfo contains runtime information about MB.
type RuntimeInfo struct {
	StartTime      time.Time `json:"startTime"`
	CWD            string    `json:"CWD"`
	GoroutineCount int       `json:"goroutineCount"`
	GOMAXPROCS     int       `json:"GOMAXPROCS"`
	GOGC           string    `json:"GOGC"`
	GODEBUG        string    `json:"GODEBUG"`
}
