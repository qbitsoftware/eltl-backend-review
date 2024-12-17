package model

import "gorm.io/gorm"

type Participant struct {
	gorm.Model
	TournamentID uint
	PlayerID     uint
}

