package model

type Tabel struct {
	Rounds       []Round             `json:"rounds"`
	TabelMatches []TabelMatch        `json:"matches"`
	Contestants  map[uint]Contestant `json:"contestants"`
	Name         string              `json:"name"`
}

type TabelArray struct {
	Tables []Tabel `json:"tables"`
}

type Round struct {
	Name string `json:"name"`
}

type TabelMatch struct {
	MatchId       uint   `json:"matchId"`
	RoundIndex    int    `json:"roundIndex"`
	Order         int    `json:"order"`
	Sides         []Side `json:"sides"`
	IsBronzeMatch bool   `json:"isBronzeMatch"`
	Bracket       string `json:"bracket"`
}

type Side struct {
	ContestantID string  `json:"contestantId"`
	Scores       []Score `json:"scores"`
	IsWinner     bool    `json:"isWinner,omitempty"` // omitempty handles the cases where the field might be missing
}

type Score struct {
	MainScore string `json:"mainScore"`
	IsWinner  bool   `json:"isWinner,omitempty"`
}

type Contestant struct {
	EntryStatus string   `json:"entryStatus"`
	Players     []Player `json:"players"`
}

type Player struct {
	Title       string `json:"title"`
	Nationality string `json:"nationality"`
}

type TournamentTabel struct {
	Teams []TeamWithMatches `json:"teams"`
}

type TeamWithMatches struct {
	Team        Team            `json:"team"`
	Matches     []MatchWithSets `json:"matches"`
	TotalPoints int             `json:"total_points"`
}

type MatchWithSets struct {
	Match          Match `json:"match"`
	Player_1_score int   `json:"player_1_score"`
	Player_2_score int   `json:"player_2_score"`
	PointsGained   int   `json:"points_gained"`
	Regrouped      bool  `json:"regrouped"`
	Sets           []Set `json:"sets"`
}
