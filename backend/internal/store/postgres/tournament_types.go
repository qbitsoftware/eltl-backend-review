package psqlstore

import (
	"fmt"
	"table-tennis/internal/model"
)

type TournamentFactory struct {
	tournament *model.Tournament
	database   *Store
	// generator  *Generator
	// phandler   *PHandler
	// visualizer *Visualizer
	// calculator *Calculator
}

func NewFactory(tournament *model.Tournament, db *Store) *TournamentFactory {
	return &TournamentFactory{
		tournament: tournament,
		database:   db,
	}
}

func (f *TournamentFactory) Generator() (Generator, error) {
	switch f.tournament.Type {
	case "single_elimination":
		return CreateGeneratorSingle(f.database), nil
	case "double_elimination":
		return CreateGeneratorDouble(f.database), nil
	case "double_elimination_final":
		return CreateGeneratorDoubleFinal(f.database), nil
	case "meistriliiga":
		return CreateGeneratorMeistriliiga(f.database), nil
	default:
		return nil, fmt.Errorf("unsupported tournament type: %s", f.tournament.Type)
	}
}

func (f *TournamentFactory) PHandler() (PHandler, error) {
	switch f.tournament.Type {
	case "single_elimination":
		return CreatePHandlerSingle(f.database), nil
	case "double_elimination":
		return CreatePHandlerDouble(f.database), nil
	case "double_elimination_final":
		return CreatePHandlerDoubleFinal(f.database), nil
	case "meistriliiga":
		return CreatePHandlerMeistriliiga(f.database), nil
	default:
		return nil, fmt.Errorf("unsupported tournament type: %s", f.tournament.Type)
	}
}

func (f *TournamentFactory) Calculator() (Calculator, error) {
	return nil, nil
}

func (f *TournamentFactory) Visualizer() (Visualizer, error) {
	return nil, nil
}
