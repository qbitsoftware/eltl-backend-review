package store

type Store interface {
	User() UserRepository
	Club() ClubRepository
	Tournament() TournamentRepository
	Match() MatchRepository
	Set() SetRepository
	Session() SessionRepository
	LoginUser() LoginUserRepository
}
