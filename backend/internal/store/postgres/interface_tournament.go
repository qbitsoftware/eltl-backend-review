package psqlstore

import (
	"table-tennis/internal/model"

	"gorm.io/gorm"
)

type Tournament interface {
	Generator() (Generator, error)
	PHandler() (PHandler, error)
	Visualizer() (Visualizer, error)
	Calculator() (Calculator, error)
}

type Generator interface {
	CreateMatches(tournament *model.Tournament, tx *gorm.DB) error
	CreateBrackets(tournament *model.Tournament, tx *gorm.DB) error
	LinkMatches(tournament *model.Tournament, tx *gorm.DB) error
	AssignPlayers(tournament *model.Tournament, tx *gorm.DB) error
}

type PHandler interface {
	MovePlayers(match *model.Match, tx *gorm.DB) error
}

type Visualizer interface {
	GetTablePlayers(tournament *model.Tournament) (map[uint]model.Contestant, error)
	GetTable(tournament *model.Tournament) (*model.TabelArray, error)
}

type Calculator interface {
	CalculateRating(tournament_id uint) error
}
