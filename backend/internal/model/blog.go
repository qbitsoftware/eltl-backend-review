package model

import "gorm.io/gorm"

type Blog struct {
	gorm.Model
	Data         string `json:"data"`
	TournamentID uint   `json:"tournament_id"`
	AuthorID     uint   `json:"authord_id"`
}
