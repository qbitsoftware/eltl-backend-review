package model

import (
	"time"

	"gorm.io/gorm"
)

type Set struct {
	gorm.Model
	MatchID    uint `json:"match_id"`
	Team1Score int  `json:"team_1_score"`
	Team2Score int  `json:"team_2_score"`
	SetNumber  int  `json:"set_number"`
}

func (s *Set) IsEqual(other Set) bool {
	return s.MatchID == other.MatchID &&
		s.Team1Score == other.Team1Score &&
		s.Team2Score == other.Team2Score &&
		s.ID == other.ID
}

//----------HOOKS------------//

func (s *Set) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed() {
		tx.Statement.SetColumn("UpdatedAt", time.Now())
	}
	return nil
}
