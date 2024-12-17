package model

import (
	"time"

	"gorm.io/gorm"
)

type Result struct {
	gorm.Model
	WinnerID  uint `json:"winner_id"`
	LoserID   uint `json:"loser_id"`
	Placement uint `json:"placement"`
}

//----------HOOKS------------//

func (r *Result) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed() {
		tx.Statement.SetColumn("UpdatedAt", time.Now())
	}
	return nil
}
