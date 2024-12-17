package model

type MatchWithTeamAndSets struct {
	MatchWithTeam MatchWithTeams  `json:"match_with_team"`
	TeamMatch     TeamMatch       `json:"team_match"`
	PlayerMatches []MatchWithSets `json:"player_matches"`
}
