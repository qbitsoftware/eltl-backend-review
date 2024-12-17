package psqlstore

import (
	"table-tennis/internal/store"

	"gorm.io/gorm"
)

type Store struct {
	Db                   *gorm.DB
	userRepository       *UserRepository
	clubRepository       *ClubRepository
	tournamentRepository *TournamentRepository
	matchRepository      *MatchRepository
	setRepository        *SetRepository
	sessionRepository    *SessionRepository
	userLoginRepository  *UserLoginRepository
}

func New(db *gorm.DB) *Store {
	return &Store{
		Db: db,
	}
}

func (s *Store) Session() store.SessionRepository {
	if s.sessionRepository != nil {
		return s.sessionRepository
	}
	s.sessionRepository = &SessionRepository{
		store: s,
	}
	return s.sessionRepository
}

func (s *Store) LoginUser() store.LoginUserRepository {
	if s.userLoginRepository != nil {
		return s.userLoginRepository
	}
	s.userLoginRepository = &UserLoginRepository{
		store: s,
	}
	return s.userLoginRepository
}

func (s *Store) User() store.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}

	s.userRepository = &UserRepository{
		store: s,
	}

	return s.userRepository
}

func (s *Store) Club() store.ClubRepository {
	if s.clubRepository != nil {
		return s.clubRepository
	}

	s.clubRepository = &ClubRepository{
		store: s,
	}

	return s.clubRepository
}

func (s *Store) Tournament() store.TournamentRepository {
	if s.tournamentRepository != nil {
		return s.tournamentRepository
	}

	s.tournamentRepository = &TournamentRepository{
		store: s,
	}

	return s.tournamentRepository
}

func (s *Store) Match() store.MatchRepository {
	if s.matchRepository != nil {
		return s.matchRepository
	}

	s.matchRepository = &MatchRepository{
		store: s,
	}

	return s.matchRepository
}

func (s *Store) Set() store.SetRepository {
	if s.setRepository != nil {
		return s.setRepository
	}

	s.setRepository = &SetRepository{
		store: s,
	}

	return s.setRepository
}
