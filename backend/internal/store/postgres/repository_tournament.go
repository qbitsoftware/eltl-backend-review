package psqlstore

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"table-tennis/internal/model"
	"table-tennis/internal/store/utils"
	"time"

	"gorm.io/gorm"
)

type TournamentRepository struct {
	store *Store
}

func (t *TournamentRepository) Create(tournament model.Tournament) error {
	tournament.State = "created"
	if result := t.store.Db.Create(&tournament); result.Error != nil {
		return result.Error
	}
	return nil
}

func (t *TournamentRepository) Delete(tournament_id uint) error {
	if err := t.store.Db.Unscoped().Where("id = ?", tournament_id).Delete(&model.Tournament{}).Error; err != nil {
		return err
	}
	return nil
}

func (t *TournamentRepository) GetWithPrivate(id uint, isLoggedIn bool) (*model.Tournament, error) {
	var tournament model.Tournament
	if isLoggedIn {
		if result := t.store.Db.First(&tournament, "id = ?", id); result.Error != nil {
			return nil, result.Error
		}
	} else {
		if err := t.store.Db.First(&tournament, "id = ? AND private = ?", id, false).Error; err != nil {
			return nil, err
		}
	}
	test := tournament.StartDate.Local().UTC()
	tournament.StartDate = &test
	test1 := tournament.EndDate.Local().UTC()
	tournament.EndDate = &test1
	return &tournament, nil
}

func (t *TournamentRepository) Get(id uint) (*model.Tournament, error) {
	var tournament model.Tournament
	if result := t.store.Db.First(&tournament, "id = ?", id); result.Error != nil {
		return nil, result.Error
	}
	test := tournament.StartDate.Local().UTC()
	tournament.StartDate = &test
	test1 := tournament.EndDate.Local().UTC()
	tournament.EndDate = &test1
	return &tournament, nil
}

func (t *TournamentRepository) Update(tournament model.Tournament) error {
	tx := t.store.Db.Begin()
	fmt.Printf("tournament coming in %+v\n", tournament)
	var existingTournament model.Tournament
	if err := tx.Where("id = ?", tournament.ID).First(&existingTournament).Error; err != nil {
		tx.Rollback()
		return err
	}

	if existingTournament.IsEqual(tournament) {
		tx.Rollback()
		return fmt.Errorf("no changes detected, nothing to update")
	}
	updateFields := map[string]interface{}{
		"Name":      tournament.Name,
		"StartDate": tournament.StartDate,
		"EndDate":   tournament.EndDate,
		"Type":      tournament.Type,
		"State":     tournament.State,
		"Private":   tournament.Private,
	}
	result := tx.Model(&model.Tournament{}).Where("id = ?", tournament.ID).Updates(updateFields)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	return tx.Commit().Error
}

func (t *TournamentRepository) CheckIfFinised(tournament_id uint, tx *gorm.DB) (bool, error) {
	var output []model.Match
	if err := tx.Model(&model.Match{}).Where("tournament_id = ? AND winner_id = 0", tournament_id).Find(&output).Error; err != nil {
		return false, err
	}

	if len(output) >= 1 {
		return false, nil
	}
	return true, nil
}

func (t *TournamentRepository) GetAll(loggedIn bool) ([]model.Tournament, error) {
	var tournaments []model.Tournament
	if loggedIn {
		if result := t.store.Db.Find(&tournaments); result.Error != nil {
			return nil, result.Error
		}
	} else {
		fmt.Println("executing this")
		if result := t.store.Db.Model(model.Tournament{}).Where("private = ?", false).Find(&tournaments); result.Error != nil {
			return nil, result.Error
		}
	}
	return tournaments, nil
}

func (t *TournamentRepository) GetRegroupedMatches(id uint) ([]model.Team, error) {

	all_teams_regrouped, err := t.store.Club().GetTournamentTeamsRegroupted(id, nil)
	if err != nil {
		return nil, err
	}

	return all_teams_regrouped, nil
}

func (t *TournamentRepository) GetParticipants(id uint) ([]model.User, error) {
	var users []model.User
	if err := t.store.Db.Model(&model.User{}).Joins("left join participants on participants.player_id = users.id").Where("participants.tournament_id = ?", id).Order("users.rating_points desc").Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (t *TournamentRepository) DeleteRegroupedMatches(id uint) error {
	if err := t.store.Db.Unscoped().Where("tournament_id = ?", id).Where("type = 'voor'").Where("current_round >= 8").Delete(&model.Match{}).Error; err != nil {
		return err
	}
	return nil
}

func (t *TournamentRepository) CreateRegroupedMatches(tournament_id uint, all_teams_regrouped []model.Team) error {
	tx := t.store.Db.Begin()

	var tournament model.Tournament
	if err := tx.Model(model.Tournament{}).Where("id = ?", tournament_id).Find(&tournament).Error; err != nil {
		return err
	}

	//check if there are any mathces remaining
	matches, err := t.store.Match().GetRemainingMeistriliigaMatches(tournament_id, tx)
	if err != nil {
		return err
	}
	if len(matches) != 0 {
		return errors.New("kõik mängud pole läbi")
	}

	//it always starts from the third day
	gameDay := 3
	//we have already generated 7 rounds
	roundNumber := 8

	for i, round := range reversePairs(MeistriLiigaMatches) {
		if i == 1 || i == 4 {
			gameDay++
		}
		for index, match := range round {
			year, month, day := tournament.StartDate.Date()
			var start_date time.Time
			if roundNumber == 8 || roundNumber == 10 || roundNumber == 13 {
				start_date = time.Date(year, month, day+gameDay-1, 13, 0, 0, 0, tournament.StartDate.Location())
			} else if roundNumber == 9 || roundNumber == 12 {
				start_date = time.Date(year, month, day+gameDay-1, 10, 0, 0, 0, tournament.StartDate.Location())
			} else if roundNumber == 11 || roundNumber == 14 {
				start_date = time.Date(year, month, day+gameDay-1, 16, 0, 0, 0, tournament.StartDate.Location())
			}
			played_match, err := t.store.Match().GetTeamVSTeamMatch(tournament_id, all_teams_regrouped[match[0]-1].ID, all_teams_regrouped[match[1]-1].ID, tx)
			if err != nil {
				return err
			}
			var matchToCreate model.Match
			if played_match.ID == 0 {
				matchToCreate = model.Match{
					TournamentID: tournament_id,
					P1ID:         all_teams_regrouped[match[0]-1].ID,
					P2ID:         all_teams_regrouped[match[1]-1].ID,
					Type:         "voor",
					CurrentRound: roundNumber, //Vooru nr
					Identifier:   gameDay,     //Game day number
					StartDate:    start_date,
					Table:        index + 1,
				}
			} else {
				matchToCreate = model.Match{
					TournamentID: tournament_id,
					P1ID:         all_teams_regrouped[match[1]-1].ID,
					P2ID:         all_teams_regrouped[match[0]-1].ID,
					Type:         "voor",
					CurrentRound: roundNumber, //Vooru nr
					Identifier:   gameDay,     //Game day number
					StartDate:    start_date,
					Table:        index + 1,
				}
			}
			//check if htere has been any previous matches
			if err := tx.Create(&matchToCreate).Error; err != nil {
				return err
			}
		}
		roundNumber++
	}
	return tx.Commit().Error
}

func (t *TournamentRepository) Generate(id uint) error {
	tournament, err := t.Get(id)
	if err != nil {
		return err
	}

	factory := NewFactory(tournament, t.store)
	generator, err := factory.Generator()
	if err != nil {
		fmt.Println("Error", err)
		return err
	}

	tx := t.store.Db.Begin()
	if err := generator.CreateMatches(tournament, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := generator.CreateBrackets(tournament, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := generator.LinkMatches(tournament, tx); err != nil {
		tx.Rollback()
		return err
	}
	if err := generator.AssignPlayers(tournament, tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (t *TournamentRepository) FinishMatch(match *model.Match, tx *gorm.DB) error {
	tournament, err := t.Get(match.TournamentID)
	if err != nil {
		return err
	}
	factory := NewFactory(tournament, t.store)
	player_handler, err := factory.PHandler()
	if err != nil {
		return err
	}
	//start the transaction
	if tx == nil {
		tx = t.store.Db.Begin()
	}
	if err := player_handler.MovePlayers(match, tx); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

func (t *TournamentRepository) GetTimeTable(id uint) ([]model.MatchWithTeams, error) {
	var matches []model.MatchWithTeams
	if err := t.store.Db.Model(&model.Match{}).
		Select(`
			matches.id,
			matches.tournament_id,
			matches.p1_id,
			matches.p2_id,
			matches.type,
			matches.current_round,
			matches.identifier,
			matches.winner_id,
			matches.bracket,
			matches.winner_next_match,
			matches.loser_next_match,
			matches.start_date,
			matches.table,
			matches.team_match_id,
			matches.head_referee,
			matches.table_referee,
			matches.place,
			matches.forfeit_match,
			team1.name as p1_team_name,
			team2.name as p2_team_name,
			team1.team_order as p1_team_order,
			team2.team_order as p2_team_order
		`).
		Joins("inner join teams as team1 on team1.id = matches.p1_id").
		Joins("inner join teams as team2 on team2.id = matches.p2_id").
		Where("matches.tournament_id = ? and type = 'voor'", id).
		Order("matches.current_round ASC").
		Order(`
			CASE
				WHEN (
					SELECT COUNT(*) FROM matches AS m WHERE m.tournament_id = matches.tournament_id
				) <= 28 THEN team1.team_order
				ELSE matches.table
			END ASC
		`).
		Scan(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

func (t *TournamentRepository) GetProtocols(tournamentID uint) ([]model.MatchWithTeamAndSets, error) {
	var output []model.MatchWithTeamAndSets
	var matches []model.MatchWithTeams
	if err := t.store.Db.Model(&model.Match{}).
		Select(`
			matches.id,
			matches.tournament_id,
			matches.p1_id,
			matches.p2_id,
			matches.type,
			matches.current_round,
			matches.identifier,
			matches.winner_id,
			matches.bracket,
			matches.winner_next_match,
			matches.loser_next_match,
			matches.start_date,
			matches.table,
			matches.team_match_id,
			matches.head_referee,
			matches.table_referee,
			matches.place,
			matches.forfeit_match,
			team1.name as p1_team_name,
			team2.name as p2_team_name,
			team1.team_order as p1_team_order,
			team2.team_order as p2_team_order
		`).
		Joins("inner join teams as team1 on team1.id = matches.p1_id").
		Joins("inner join teams as team2 on team2.id = matches.p2_id").
		Where("matches.tournament_id = ? AND matches.winner_id != 0", tournamentID).
		Order("matches.start_date ASC").
		Scan(&matches).Error; err != nil {
		return nil, err
	}

	for _, match := range matches {
		teamMatch, err := t.store.Club().GetTeamMatch(match.ID)
		if err != nil {
			return nil, err
		}
		// allTeamMatches, err := t.store.Club().GetAllTeamMatches(tournamentID, teamMatch.ID)
		// if err != nil {
		// 	return nil, err
		// }
		allTeamMatches, err := t.store.Club().GetTeamMatches(tournamentID, teamMatch.ID)
		if err != nil {
			return nil, err
		}
		var matchToAppend []model.MatchWithSets
		for _, tmatch := range allTeamMatches {
			var p1points int
			var p2points int
			sets, err := t.store.Match().GetMatchSets(tmatch.ID)
			if err != nil {
				return nil, err
			}
			for _, set := range sets {
				if set.Team1Score > set.Team2Score {
					p1points++
				} else if set.Team2Score > set.Team1Score {
					p2points++
				}
			}
			matchToAppend = append(matchToAppend, model.MatchWithSets{
				Match:          tmatch,
				Player_1_score: p1points,
				Player_2_score: p2points,
				Sets:           sets,
			})
		}

		output = append(output, model.MatchWithTeamAndSets{
			MatchWithTeam: match,
			TeamMatch:     *teamMatch,
			PlayerMatches: matchToAppend,
		})
	}

	return output, nil
}

func (t *TournamentRepository) ChangeMatchTime(toChange []model.ChangeClock) error {
	for _, clock := range toChange {
		var matches []model.Match

		err := t.store.Db.Where("tournament_id = ? AND current_round = ? AND type = ?", clock.TournamentID, clock.Round, "voor").Find(&matches).Error
		if err != nil {
			return err
		}
		for _, match := range matches {
			parsedTime, err := time.Parse(time.RFC3339, clock.Time)
			if err != nil {
				return fmt.Errorf("invalid time format for match %d: %v", match.Identifier, err)
			}
			newStartDate := time.Date(
				parsedTime.Year(),
				parsedTime.Month(),
				parsedTime.Day(),
				parsedTime.Hour(),
				parsedTime.Minute(),
				0, 0, // Seconds and nanoseconds
				parsedTime.UTC().Location(),
			)

			match.StartDate = newStartDate
			match.Place = clock.Place

			err = t.store.Db.Save(&match).Error
			if err != nil {
				return fmt.Errorf("failed to update match %d: %v", match.Identifier, err)
			}
		}
	}
	return nil
}

func (t *TournamentRepository) CreateBracketRecursive(tournament *model.Tournament, from, to, round int, tx *gorm.DB) error {
	upComingMatches := (to - from + 1) / 2
	// stop the recursion
	if upComingMatches == 1 {
		//create a last placement match
		match := model.Match{
			TournamentID:    tournament.ID,
			P1ID:            0,
			P2ID:            0,
			Type:            "bracket",
			CurrentRound:    round + 1,
			Identifier:      2,
			WinnerID:        0,
			Bracket:         fmt.Sprintf("%d-%d", from, to),
			WinnerNextMatch: 0,
			LoserNextMatch:  0,
		}
		if err := tx.Create(&match).Error; err != nil {
			tx.Rollback()
			return err
		}

		return nil
	}

	rounds := utils.CalculateMatchesPerRoundSingle(upComingMatches * 2)

	for i := 1; i <= len(rounds); i++ {
		for j := 1; j <= rounds[i]; j++ {
			bracket := fmt.Sprintf("%d-%d", from, from+(2*rounds[i])-1)
			if i == len(rounds) && j == rounds[i] {
				bracket = fmt.Sprintf("%d-%d", from, from+1)
			}
			match := model.Match{
				TournamentID: tournament.ID,
				P1ID:         0,
				P2ID:         0,
				Type:         "bracket",
				CurrentRound: i,
				Identifier:   j,
				WinnerID:     0,
				Bracket:      bracket,
			}
			if err := tx.Create(&match).Error; err != nil {
				tx.Rollback()
				return err
			}

		}
		if rounds[i] >= 2 {
			if err := t.CreateBracketRecursive(tournament, from+rounds[i], from+(2*rounds[i])-1, i, tx); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return nil
}

func (t *TournamentRepository) GetBrackets(tournament *model.Tournament, tx *gorm.DB) (map[int]string, error) {
	maxRounds, err := t.store.Match().GetLoserMaxRounds(tournament.ID, tx)
	output := make(map[int]string)
	if err != nil {
		return nil, err
	}

	//first of all create bracket
	lastRoundParticipants, err := t.GetParticipants(tournament.ID)
	if err != nil {
		return nil, err
	}

	last_participants := utils.RoundUpToPowerOf2(len(lastRoundParticipants))
	for i := 1; i < maxRounds; i++ {
		participants, err := t.store.Match().GetLoserCurrentRoundParticipants(tournament.ID, i, tx)
		if err != nil {
			return nil, err
		}

		until := last_participants - (participants/2 - 1)

		output[i] = fmt.Sprintf("%d-%d", until, last_participants)
		last_participants = until - 1
	}

	return output, nil
}

func (t *TournamentRepository) CalculateRating(tournament_id uint) error {
	allMatches, err := t.store.Match().GetAllTournamentMatches(tournament_id)
	if err != nil {
		return err
	}
	players, err := t.GetParticipants(tournament_id)
	if err != nil {
		return err
	}
	playerMap := make(map[uint]*model.User)
	//create a map from players so its more efficient to later find and modify the map
	for _, p := range players {
		playerMap[p.ID] = &p
	}
	//loop over allmatches and calc weights
	for _, match := range allMatches {
		//calculate Hv and Hk
		//lets make so that p1 is always winner in our context
		var p1 *model.User
		var p2 *model.User
		var Hv int
		var Hk int
		if match.P1ID == math.MaxUint32 || match.P2ID == math.MaxUint32 {
			continue
		}
		if match.WinnerID == match.P1ID {
			p1 = playerMap[match.P1ID]
			p2 = playerMap[match.P2ID]
		} else {
			p1 = playerMap[match.P2ID]
			p2 = playerMap[match.P1ID]
		}
		//if player with higher rating wins the match we have a table where we look the numbers
		Rv := math.Abs(float64(p1.RatingPoints - p2.RatingPoints))
		//Erandid
		if p1.RatingPoints == 0 || p2.RatingPoints == 0 {
			Hv = 0
			Hk = 0
		} else if p1.RatingPoints > p2.RatingPoints {
			if Rv >= 0 && Rv <= 2 {
				Hv = 2
				Hk = -2
			} else if Rv >= 3 && Rv <= 13 {
				Hv = 1
				Hk = -1
			} else {
				Hv = 0
				Hk = 0
			}
		} else {
			Hv = int((Rv + 5) / 3)
			Hk = Hv * (-1)
		}
		p1.Hv += Hv
		p2.Hk += Hk
	}
	//Calculate RP
	for _, value := range playerMap {
		divier := (value.WeightPoints + value.Hv + value.Hk)
		if divier == 0 {
			divier = 1
		}
		RpToAdd := (((value.Hv - value.Hk) * 10) + value.Hv*1) / divier
		value.RatingPoints += RpToAdd
	}

	//update players in database
	for _, player := range playerMap {
		if err := t.store.Db.Model(&model.User{}).Where("id = ?", player.ID).Update("rating_points", player.RatingPoints).Error; err != nil {
			fmt.Println("error updating rating points", err)
			return err
		}
	}
	fmt.Println("Successfully updated rating points")
	return nil
}

func (t *TournamentRepository) ShowTable(id uint) ([]model.TabelArray, error) {
	var output []model.TabelArray
	tournament, err := t.Get(id)
	if err != nil {
		return nil, err
	}
	switch tournament.Type {
	case "single_elimination":
		tabel, err := t.GetTabel(tournament, "winner")
		if err != nil {
			return nil, err
		}
		output = []model.TabelArray{*tabel}
	case "double_elimination":
		winner, err := t.GetTabel(tournament, "winner")
		if err != nil {
			return nil, err
		}
		winner.Tables[0].Name = "Plussring"
		loser, err := t.GetTabel(tournament, "loser")
		if err != nil {
			return nil, err
		}
		loser.Tables[0].Name = "Miinusring"
		brackets, err := t.GetBracketsTabel(tournament)
		if err != nil {
			return nil, err
		}
		output = []model.TabelArray{*winner, *loser}
		output = append(output, brackets...)
	case "double_elimination_final":
		winner, err := t.GetTabel(tournament, "winner")
		if err != nil {
			return nil, err
		}
		winner.Tables[0].Name = "Plussring"
		loser, err := t.GetTabel(tournament, "loser")
		if err != nil {
			return nil, err
		}
		loser.Tables[0].Name = "Miinusring"
		brackets, err := t.GetBracketsTabel(tournament)
		if err != nil {
			return nil, err
		}
		output = []model.TabelArray{*winner, *loser}
		output = append(output, brackets...)
	case "meistriliiga":
		//implement meistriliiga

		fmt.Println("Trying to get results to meistriliiga...")
	default:
		return nil, errors.New("no table for provided tournament")
	}
	return output, nil
}

func (t *TournamentRepository) GetMeistriliigaTabel(tournament_id uint) (*model.TournamentTabel, error) {
	var output model.TournamentTabel
	tournament, err := t.Get(tournament_id)
	if err != nil {
		return nil, err
	}
	teams, err := t.store.Club().GetTournamentTeams(tournament.ID)
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		var teamStruct model.TeamWithMatches
		teamMatches, err := t.store.Club().GetAllTeamMatches(tournament.ID, team.ID)
		if err != nil {
			return nil, err
		}

		for _, match := range teamMatches {
			team_match, err := t.store.Club().GetTeamMatch(match.ID)
			if err != nil {
				if err != gorm.ErrRecordNotFound {
					// fmt.Println("JEP This is expected", err)
					continue
				} else {
					return nil, err
				}
			}

			team_matches, err := t.store.Club().GetTeamMatches(tournament_id, team_match.ID)
			if err != nil {
				return nil, err
			}
			p1_points := 0
			p2_points := 0
			points_gained := 0

			for _, t_match := range team_matches {
				var p1points int
				var p2points int
				sets, err := t.store.Match().GetMatchSets(t_match.ID)
				if err != nil {
					return nil, err
				}
				for _, set := range sets {
					if set.Team1Score > set.Team2Score {
						p1points++
					} else if set.Team2Score > set.Team1Score {
						p2points++
					}
				}
				if p1points > p2points {
					p1_points++
				} else if p2points > p1points {
					p2_points++
				}

			}
			var regrouped bool
			if match.CurrentRound >= 8 {
				regrouped = true
			}
			if team.ID == match.P1ID && p1_points > p2_points {
				points_gained = 2
			} else if team.ID == match.P2ID && p2_points > p1_points {
				points_gained = 2
			} else if p1_points == 0 && p2_points == 0 {
				points_gained = 0
			} else {
				points_gained = 1
			}
			if match.WinnerID == 0 {
				points_gained = 0
			}
			matchStruct := model.MatchWithSets{
				Match:          match,
				Player_1_score: p1_points,
				Player_2_score: p2_points,
				PointsGained:   points_gained,
				Regrouped:      regrouped,
			}
			teamStruct.TotalPoints += points_gained
			teamStruct.Matches = append(teamStruct.Matches, matchStruct)
		}
		teamStruct.Team = team
		output.Teams = append(output.Teams, teamStruct)
	}
	return &output, nil
}

func (t *TournamentRepository) GetBracketsTabel(tournament *model.Tournament) ([]model.TabelArray, error) {
	var finalOutput []model.TabelArray

	brackets, err := t.GetBrackets(tournament, t.store.Db)
	if err != nil {
		return nil, err
	}

	contestans, err := t.GetTablePlayers(tournament)
	if err != nil {
		return nil, err
	}

	for _, value := range brackets {
		var output []model.Tabel
		bracket, err := utils.ParseBracketString(value)
		if err != nil {
			return nil, err
		}

		matchesToDisplay, err := t.store.Match().GetAllBracketMatchesWeak(tournament, bracket[0], bracket[1], t.store.Db)
		if err != nil {
			return nil, err
		}
		var tabelMatch []model.TabelMatch

		//find breakboint for displaying purposes
		var lastRound int
		for index, match := range matchesToDisplay {
			//find scores
			p1_score, err := t.store.Match().GetPlayerScores(match.ID)
			if err != nil {
				return nil, err
			}
			p1Scores := []model.Score{}
			p2Scores := []model.Score{}
			for _, set := range p1_score {
				p1Scores = append(p1Scores, model.Score{
					MainScore: strconv.Itoa(set.Team1Score),
				})
			}
			for _, set := range p1_score {
				p2Scores = append(p2Scores, model.Score{
					MainScore: strconv.Itoa(set.Team2Score),
				})
			}
			//check if its the last match in this bracket
			if index+1 < len(matchesToDisplay) {

				//this is the last match of this bracket
				if matchesToDisplay[index+1].CurrentRound < lastRound {
					//add finale game
					tabelMatch = append(tabelMatch, model.TabelMatch{
						MatchId:       match.ID,
						RoundIndex:    match.CurrentRound - 1,
						Order:         1,
						IsBronzeMatch: true,
						Sides: []model.Side{
							{
								ContestantID: strconv.FormatUint(uint64(match.P1ID), 10),
								Scores:       p1Scores,
							},
							{
								ContestantID: strconv.FormatUint(uint64(match.P2ID), 10),
								Scores:       p2Scores,
							},
						},
						Bracket: match.Bracket,
					})

					rounds := t.GetTabelRounds(tabelMatch)
					output = append(output, model.Tabel{
						Rounds:       rounds,
						TabelMatches: tabelMatch,
						Contestants:  contestans,
						Name:         fmt.Sprintf("Kohad %d-%d", bracket[0], bracket[1]),
					})
					tabelMatch = []model.TabelMatch{}
					lastRound = 1
					continue
				}
			} else if index+1 >= len(matchesToDisplay) {
				//add finale game
				roundIndex := match.CurrentRound - 1
				finalMatch := true
				order := 1
				if bracket[0] == 5 && bracket[1] == 6 {
					roundIndex = 0
					finalMatch = false
					order = 0
				} else if bracket[0] == 7 && bracket[1] == 8 {
					roundIndex = 0
					finalMatch = false
					order = 0
				}
				tabelMatch = append(tabelMatch, model.TabelMatch{
					MatchId:       match.ID,
					RoundIndex:    roundIndex,
					Order:         order,
					IsBronzeMatch: finalMatch,
					Sides: []model.Side{
						{
							ContestantID: strconv.FormatUint(uint64(match.P1ID), 10),
							Scores:       p1Scores,
						},
						{
							ContestantID: strconv.FormatUint(uint64(match.P2ID), 10),
							Scores:       p2Scores,
						},
					},
					Bracket: match.Bracket,
				})
				rounds := t.GetTabelRounds(tabelMatch)
				output = append(output, model.Tabel{
					Rounds:       rounds,
					TabelMatches: tabelMatch,
					Contestants:  contestans,
					Name:         fmt.Sprintf("Kohad %d-%d", bracket[0], bracket[1]),
				})
				break
			}
			//else just add to normal

			tabelMatch = append(tabelMatch, model.TabelMatch{
				MatchId:    match.ID,
				RoundIndex: match.CurrentRound - 1,
				Order:      match.Identifier - 1,
				Sides: []model.Side{
					{
						ContestantID: strconv.FormatUint(uint64(match.P1ID), 10),
						Scores:       p1Scores,
					},
					{
						ContestantID: strconv.FormatUint(uint64(match.P2ID), 10),
						Scores:       p2Scores,
					},
				},
				Bracket: match.Bracket,
			})
			lastRound = match.CurrentRound
		}

		finalOutput = append(finalOutput, model.TabelArray{Tables: output})

	}

	return finalOutput, nil
}

func (t *TournamentRepository) GetTabel(tournament *model.Tournament, match_type string) (*model.TabelArray, error) {
	var tabel model.TabelArray

	matches, err := t.GetTabelMatches(*tournament, match_type)
	if err != nil {
		return nil, err
	}

	contestans, err := t.GetTablePlayers(tournament)
	if err != nil {
		return nil, err
	}

	rounds := t.GetTabelRounds(matches)

	tabel.Tables = append(tabel.Tables, model.Tabel{
		TabelMatches: matches,
		Contestants:  contestans,
		Rounds:       rounds,
	})

	return &tabel, nil
}

func (t *TournamentRepository) GetTabelRounds(matches []model.TabelMatch) []model.Round {
	var output []model.Round
	var maxRounds int
	for _, m := range matches {
		if m.RoundIndex > maxRounds {
			maxRounds = m.RoundIndex
		}
	}

	for i := 0; i <= maxRounds; i++ {
		name := ""
		if maxRounds-i == 0 {
			name = ""
		} else if maxRounds-i == 1 {
			name = ""
		}
		output = append(output, model.Round{
			Name: name,
		})
	}
	return output
}

func (t *TournamentRepository) GetTablePlayers(tournament *model.Tournament) (map[uint]model.Contestant, error) {
	participants, err := t.GetParticipants(tournament.ID)
	if err != nil {
		return nil, err
	}
	output := make(map[uint]model.Contestant)
	for i := 0; i < len(participants); i++ {
		contestant := model.Contestant{}

		contestant.EntryStatus = strconv.Itoa(i + 1)
		contestant.Players = []model.Player{
			{
				Title: participants[i].FirstName + " " + participants[i].LastName,
			},
		}

		output[participants[i].ID] = contestant
	}

	byebye := model.Contestant{}
	byebye.EntryStatus = ""
	byebye.Players = []model.Player{
		{
			Title: "bye-bye",
		},
	}
	output[math.MaxUint32] = byebye
	return output, nil
}

func (t *TournamentRepository) GetTabelMatches(tournament model.Tournament, match_type string) ([]model.TabelMatch, error) {
	var output []model.TabelMatch

	m, err := t.store.Match().GetTournamentMatches(tournament.ID, match_type)
	if err != nil {
		return nil, err
	}
	// if there are no matches we know that those are not generated then just return nothing
	if len(m) <= 0 {
		return nil, errors.New("no matches on this tournament yet")
	}

	for _, match := range m {
		p1_score, err := t.store.Match().GetPlayerScores(match.ID)
		if err != nil {
			return nil, err
		}
		p1Scores := []model.Score{}
		p2Scores := []model.Score{}
		for _, set := range p1_score {
			p1Scores = append(p1Scores, model.Score{
				MainScore: strconv.Itoa(set.Team1Score),
			})
		}
		for _, set := range p1_score {
			p2Scores = append(p2Scores, model.Score{
				MainScore: strconv.Itoa(set.Team2Score),
			})
		}

		finalMatch := false
		if tournament.Type == "double_elimination_final" && match.Bracket == "3-4" {
			finalMatch = true
		}

		output = append(output, model.TabelMatch{
			MatchId:       match.ID,
			RoundIndex:    match.CurrentRound - 1,
			Order:         match.Identifier - 1,
			IsBronzeMatch: finalMatch,
			Sides: []model.Side{
				{
					ContestantID: strconv.FormatUint(uint64(match.P1ID), 10),
					Scores:       p1Scores,
				},
				{
					ContestantID: strconv.FormatUint(uint64(match.P2ID), 10),
					Scores:       p2Scores,
				},
			},
			Bracket: match.Bracket,
		})
	}
	return output, nil
}
