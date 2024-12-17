package store

import (
	"table-tennis/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	Create(model.User) error
	Get(uint) (*model.User, error)
	GetPlayers(mId uint) ([]model.User, error)
	GetAll() ([]model.User, error)
	GetMatches(user_id uint) ([]model.Match, error)
	ScrapeUsers() error
	GetTournamentBlog(tournament_id uint) (*model.Blog, error)
	CreateBlog(blog model.Blog) error
}

type LoginUserRepository interface {
	GetByLogin(email string) (*model.LoginUser, error)
	Get(id uint) (*model.LoginUser, error)
	Create(user model.LoginUser) error
}

type ClubRepository interface {
	CreateTeam(model.Team) error
	UpdateTeam(model.Team) error
	GetTournamentTeams(uint) ([]model.Team, error)
	UpdateTeams(teams []model.Team, tournament_id uint) error
	DeleteTeam(team_id uint, tournament_id uint) error
	GetTournamentTeam(teamID uint) (*model.Team, error)
	GetTeamMatch(matchID uint) (*model.TeamMatch, error)
	CreateTeamMatch(teamMatch model.TeamMatch) (*model.TeamMatch, error)
	UpdateTeamMatch(teamMatch model.TeamMatch) error
	GetTeamMatches(tournament_id, team_match_id uint) ([]model.Match, error)
	GetAllTeamMatches(tournament_id, team_id uint) ([]model.Match, error)
	DeleteTeamMatch(team_match_id uint) error
	GetTournamentTeamsRegroupted(tournamentID uint, tx *gorm.DB) ([]model.Team, error)
}

type TournamentRepository interface {
	Create(model.Tournament) error
	Get(uint) (*model.Tournament, error)
	GetWithPrivate(id uint, isLoggedIn bool) (*model.Tournament, error)
	Update(model.Tournament) error
	GetAll(loggedIn bool) ([]model.Tournament, error)
	GetTimeTable(id uint) ([]model.MatchWithTeams, error)
	GetParticipants(id uint) ([]model.User, error)
	CheckIfFinised(tournament_id uint, tx *gorm.DB) (bool, error)
	CalculateRating(tournament_id uint) error
	ShowTable(id uint) ([]model.TabelArray, error)
	CreateBracketRecursive(tournament *model.Tournament, from, to, round int, tx *gorm.DB) error
	GetBrackets(tournament *model.Tournament, tx *gorm.DB) (map[int]string, error)
	Generate(id uint) error
	FinishMatch(match *model.Match, tx *gorm.DB) error
	ChangeMatchTime(updates []model.ChangeClock) error
	GetMeistriliigaTabel(tournament_id uint) (*model.TournamentTabel, error)
	GetProtocols(tournamentID uint) ([]model.MatchWithTeamAndSets, error)
	Delete(tournament_id uint) error
	GetRegroupedMatches(id uint) ([]model.Team, error)
	CreateRegroupedMatches(tournament_id uint, all_teams_regrouped []model.Team) error
	DeleteRegroupedMatches(id uint) error
}

type MatchRepository interface {
	CreateMatch(match model.Match) error
	GetRoundMatches(round int, tid uint, tx *gorm.DB, state string) ([]model.Match, error)
	GetTournamentMatches(tid uint, t_type string) ([]model.Match, error)
	GetPlayerScores(matchID uint) ([]model.Set, error)
	GetMatchSets(matchID uint) ([]model.Set, error)
	GetLoserCurrentRoundParticipants(tid uint, currentRound int, tx *gorm.DB) (int, error)
	GetLoserMaxRounds(tid uint, tx *gorm.DB) (int, error)
	GetAllWinnerMatches(tournament *model.Tournament, tx *gorm.DB) ([]model.Match, error)
	GetAllLoserMatches(tournament *model.Tournament, tx *gorm.DB) ([]model.Match, error)
	GetAllBracketMatches(tournament *model.Tournament, from, to int, tx *gorm.DB) ([]model.Match, error)
	GetAllBracketMatchesWeak(tournament *model.Tournament, from, to int, tx *gorm.DB) ([]model.Match, error)
	CreateSets(match uint, p1s int, p2s int) error
	GetAllTournamentMatches(tid uint) ([]model.Match, error)
	FindWinner(match model.Match, tx *gorm.DB) (winner *uint, loser *uint, finalMatch bool, err error)
	UpdateTable(matchID uint, table int) error
	GetRemainingMeistriliigaMatches(tid uint, tx *gorm.DB) ([]model.Match, error)
	GetTeamVSTeamMatch(tournament_id, team_1, team_2 uint, tx *gorm.DB) (*model.Match, error)
	UpdateForfeithMatch(match_id, winner_id uint, isForfeitMatch bool) error
}

type SetRepository interface {
	Create(set model.Set) error
	Update(set model.Set) error
	ResetMatchSets(matchID uint, tx *gorm.DB) error
	UpdateMatchSets(matchID uint, scores map[int][2]int, tx *gorm.DB) error
}

type SessionRepository interface {
	Check(access_id string) (*model.Session, error)
	Delete(access_id string) error
	Update(access_id string, session model.Session) (*model.Session, error)
	Create(session *model.Session) (*model.Session, error)
	DeleteByUser(user_id uint) error
	CheckByUserId(user_id uint) (*model.Session, error)
}
