package domain

import (
	"time"
)

type SystemSetting struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type SSETicket struct {
	Ticket    string    `json:"ticket"`
	Username  string    `json:"username"`
	ExpiresAt time.Time `json:"expires_at"`
}

type SSEEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}
