package model

import (
	"time"

	"gorm.io/gorm"
)

type TeamMatch struct {
	gorm.Model
	TournamentID uint   `json:"tournament_id"`
	WinnerID     uint   `json:"winner_id"`
	PlayerAID    uint   `json:"player_a_id" validate:"required"`
	PlayerBID    uint   `json:"player_b_id" validate:"required"`
	PlayerCID    uint   `json:"player_c_id" validate:"required"`
	PlayerDID    uint   `json:"player_d_id"`
	PlayerEID    uint   `json:"player_e_id"`
	PlayerXID    uint   `json:"player_x_id" validate:"required"`
	PlayerYID    uint   `json:"player_y_id" validate:"required"`
	PlayerZID    uint   `json:"player_z_id" validate:"required"`
	PlayerVID    uint   `json:"player_v_id"`
	PlayerWID    uint   `json:"player_w_id"`
	MatchID      uint   `json:"match_id" validate:"required"`
	CaptainA     string `json:"captain_a"`
	CaptainB     string `json:"captain_b"`
	Notes        string `json:"notes"`
}

//----------HOOKS------------//

func (m *TeamMatch) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed() {
		tx.Statement.SetColumn("UpdatedAt", time.Now())
	}
	return nil
}
