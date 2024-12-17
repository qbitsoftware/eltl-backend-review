package model

import "time"

type Session struct {
	AcessID   string    `json:"access_id"`
	UserID    uint      `json:"user_id"`
	CreatedAT time.Time `json:"timestamp"`
}
