package domain

import "time"

type HistoryRecord struct {
	LastConnected time.Time `json:"last_connected"`
	Count         int       `json:"count"`
}

type Server struct {
	Alias        string
	Host         string
	User         string
	Port         int
	IdentityFile string
	BlockIndex   int
	Checked      bool
	Online       bool
	LastUsed     time.Time
	UseCount     int
}
