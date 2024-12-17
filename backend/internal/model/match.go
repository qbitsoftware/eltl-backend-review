package model

import (
	"time"

	"gorm.io/gorm"
)

type Match struct {
	gorm.Model
	TournamentID    uint      `json:"tournament_id"`
	P1ID            uint      `json:"p1_id"`
	P2ID            uint      `json:"p2_id"`
	P1ID_2          uint      `json:"p1_id_2"`
	P2ID_2          uint      `json:"p2_id_2"`
	Type            string    `json:"type"`
	CurrentRound    int       `json:"current_round"`
	Identifier      int       `json:"identifier"`
	WinnerID        uint      `json:"winner_id"`
	Bracket         string    `json:"bracket"`
	StartDate       time.Time `json:"start_date"`
	Table           int       `json:"table"`
	Place           string    `json:"place"`
	WinnerNextMatch uint      `json:"winner_next_match"`
	LoserNextMatch  uint      `json:"loser_next_match"`
	HeadReferee     string    `json:"head_referee"`
	TableReferee    string    `json:"table_referee"`
	TeamMatchID     uint      `json:"team_match_id"`
	ForfeitMatch    bool      `json:"forfeit_match"`
}

//----------HOOKS------------//

func (m *Match) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed() {
		tx.Statement.SetColumn("UpdatedAt", time.Now())
	}
	return nil
}
