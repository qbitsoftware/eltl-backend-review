package model

import (
	"time"

	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	Name         string `json:"name"`
	TournamentID uint   `json:"tournament_id"`
	Captain      string `json:"captain_id"`
	TeamOrder    int    `json:"order"`
	TotalPoints  int    `gorm:"-" json:"total_points"`
	Players      []User `gorm:"-" json:"players"`
}

type PlayerJoinTeam struct {
	gorm.Model
	PlayerID     uint
	Confirmation string
	HasRating    bool
	TeamID       uint
}

type MatchWithTeams struct {
	ID              uint      `json:"id"`
	TournamentID    uint      `json:"tournament_id"`
	P1ID            uint      `json:"p1_id"`
	P2ID            uint      `json:"p2_id"`
	Type            string    `json:"type"`
	CurrentRound    int       `json:"current_round"`
	Identifier      int       `json:"identifier"`
	WinnerID        uint      `json:"winner_id"`
	Bracket         string    `json:"bracket"`
	StartDate       time.Time `json:"start_date"`
	WinnerNextMatch uint      `json:"winner_next_match"`
	LoserNextMatch  uint      `json:"loser_next_match"`
	Table           int       `json:"table"`
	TeamMatchID     uint      `json:"team_match_id"`
	HeadReferee     string    `json:"head_referee"`
	TableReferee    string    `json:"table_referee"`
	Place           string    `json:"place"`
	IsForfeitMatch  bool      `json:"forfeit_match"`

	P1TeamName  string `json:"p1_team_name"`
	P2TeamName  string `json:"p2_team_name"`
	P1TeamOrder int
	P2TeamOrder int
}

type ChangeClock struct {
	Round        int    `json:"round"`
	TournamentID int    `json:"tournament_id"`
	Time         string `json:"time"`
	Place        string `json:"place"`
}
