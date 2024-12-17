package psqlstore

import (
	"errors"
	"fmt"
	"math"
	"table-tennis/internal/model"

	"gorm.io/gorm"
)

type MatchRepository struct {
	store *Store
}

func (m *MatchRepository) CreateMatch(match model.Match) error {
	if err := m.store.Db.Create(&match).Error; err != nil {
		return err
	}
	return nil
}

func (m *MatchRepository) GetRoundMatches(round int, tid uint, tx *gorm.DB, state string) ([]model.Match, error) {
	var matches []model.Match
	if result := tx.Model(&model.Match{}).Where("tournament_id = ? AND current_round = ? AND type = ?", tid, round, state).Find(&matches); result.Error != nil {
		return nil, result.Error
	}
	return matches, nil
}

func (m *MatchRepository) GetTournamentMatches(tid uint, t_type string) ([]model.Match, error) {
	var matches []model.Match
	if err := m.store.Db.Model(&model.Match{}).Where("tournament_id = ? AND type = ?", tid, t_type).Order("current_round asc").Order("identifier asc").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}
func (m *MatchRepository) GetAllTournamentMatches(tid uint) ([]model.Match, error) {
	var matches []model.Match
	if err := m.store.Db.Model(&model.Match{}).Where("tournament_id = ?", tid).Order("current_round asc").Order("identifier asc").Find(&matches).Error; err != nil {
		return nil, err
	}
	return matches, nil
}

func (m *MatchRepository) GetPlayerScores(matchID uint) ([]model.Set, error) {
	var sets []model.Set
	if err := m.store.Db.Model(&model.Set{}).Where("match_id = ?", matchID).Find(&sets).Error; err != nil {
		return nil, err
	}
	return sets, nil
}

func (m *MatchRepository) GetMatchSets(matchID uint) ([]model.Set, error) {
	var sets []model.Set
	if err := m.store.Db.Model(&model.Set{}).Where("match_id = ?", matchID).Order("set_number ASC").Find(&sets).Error; err != nil {
		return nil, err
	}
	return sets, nil
}

func (m *MatchRepository) UpdateTable(matchID uint, table int) error {
	if err := m.store.Db.Model(&model.Match{}).Where("id = ?", matchID).Update("table", table).Error; err != nil {
		return err
	}
	return nil
}

func (m *MatchRepository) GetLoserMaxRounds(tid uint, tx *gorm.DB) (int, error) {
	var highestRound int

	if err := tx.Model(&model.Match{}).Where("tournament_id = ? AND type = 'loser'", tid).Select("MAX(current_round)").Scan(&highestRound).Error; err != nil {
		return 0, err
	}

	return highestRound, nil
}

func (m *MatchRepository) GetLoserCurrentRoundParticipants(tid uint, currentRound int, tx *gorm.DB) (int, error) {
	var output []model.Match
	if err := tx.Model(&model.Match{}).Where("tournament_id = ? AND current_round = ? AND type = 'loser'", tid, currentRound).Find(&output).Error; err != nil {
		return 0, err
	}
	return len(output) * 2, nil
}

func (m *MatchRepository) GetRemainingMeistriliigaMatches(tid uint, tx *gorm.DB) ([]model.Match, error) {
	var output []model.Match
	if tx == nil {
		tx = m.store.Db.Begin()
	}
	if err := tx.Model(&model.Match{}).Where("tournament_id = ? AND type = 'voor' AND winner_id = 0", tid).Find(&output).Error; err != nil {
		return nil, err
	}
	return output, nil
}

func (m *MatchRepository) GetFirstLoserMatch(tournamentID uint, tx *gorm.DB) (*model.Match, error) {
	var match model.Match
	if err := tx.First(&match, "type = ? AND tournament_id = ? AND (p1_id != 0 OR p2_id != 0)", "loser", tournamentID).Error; err != nil {
		return nil, err
	}
	return &match, nil
}

func (m *MatchRepository) FindWinner(match model.Match, tx *gorm.DB) (winner *uint, loser *uint, finalMatch bool, err error) {
	var sets []model.Set
	fm := false

	if err := tx.Where("match_id = ?", match.ID).Find(&sets).Error; err != nil {
		return nil, nil, fm, err
	}

	if err := tx.Model(&model.Match{}).First(&match, "id = ?", match.ID).Error; err != nil {
		return nil, nil, fm, err
	}
	if match.WinnerID != 0 {
		return nil, nil, fm, errors.New("match already has a winner")
	}

	team1Wins := 0
	team2Wins := 0

	for _, set := range sets {
		if set.Team1Score > set.Team2Score {
			team1Wins++
		} else if set.Team2Score > set.Team1Score {
			team2Wins++
		}
	}

	//check if its the placement match
	if match.LoserNextMatch == 0 && match.WinnerNextMatch == 0 {
		fm = true
	}
	if team1Wins == team2Wins && (match.P1ID != math.MaxUint32 && match.P2ID != math.MaxUint32) {
		return nil, nil, fm, errors.New("draw")
	}

	if team1Wins > team2Wins {
		return &match.P1ID, &match.P2ID, fm, nil
	}
	return &match.P2ID, &match.P1ID, fm, nil
}

func (m *MatchRepository) CreateSets(match uint, p1s int, p2s int) error {
	//create p1 sets
	for i := 1; i <= p1s; i++ {
		if err := m.store.Set().Create(model.Set{
			MatchID:    match,
			Team1Score: 11,
			Team2Score: 0,
			SetNumber:  i,
		}); err != nil {
			return err
		}
	}
	for i := 1; i <= p2s; i++ {
		if err := m.store.Set().Create(model.Set{
			MatchID:    match,
			Team1Score: 0,
			Team2Score: 11,
			SetNumber:  i,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *MatchRepository) MovePlayers(match *model.Match, tx *gorm.DB) error {
	winner, loser, fm, err := m.FindWinner(*match, tx)
	if err != nil {
		tx.Rollback()
		return err
	}
	//query the match
	if err := tx.Model(&model.Match{}).Where("id = ?", match.ID).Scan(&match).Error; err != nil {
		tx.Rollback()
		return err
	}

	//update current match winnerid
	if err := tx.Model(&match).Update("winner_id", winner).Error; err != nil {
		tx.Rollback()
		return err
	}
	//if placement match
	if fm {
		finished, err := m.store.Tournament().CheckIfFinised(match.TournamentID, tx)
		if err != nil {
			tx.Rollback()
			return err
		}
		if finished {
			if err := tx.Model(&model.Tournament{}).Where("id = ?", match.TournamentID).Update("state", "finished").Error; err != nil {
				fmt.Println("couldnt finish tournament", err)
				tx.Rollback()
				return err
			}
			fmt.Println("tournament is finished")
			fmt.Println("Starting calculating")
			// start calculating
			if err := m.store.Tournament().CalculateRating(match.TournamentID); err != nil {
				fmt.Println("Error")
				tx.Rollback()
				return err
			}
			return tx.Commit().Error
		}

		return tx.Commit().Error
	}
	//check if tournament is now over
	//query next matches
	if err := m.HandleByeByeMoving(*match, *winner, *loser, tx); err != nil {
		return err
	}

	return nil
}

//refactored shit

func (m *MatchRepository) CheckByeBye(match model.Match, tx *gorm.DB) error {
	//check if either of them are bye bye
	// true:
	//------then move
	// false:
	//------check sets
	//------move accordingly
	if tx == nil {
		tx = m.store.Db.Begin()
	}

	if err := tx.Model(&model.Match{}).First(&match, "id = ?", match.ID).Error; err != nil {
		return err
	}
	if match.WinnerID != 0 {
		return nil
	}

	// if first player  is byebye
	if match.P1ID == math.MaxUint32 && match.P2ID != 0 && match.P2ID != math.MaxUint32 {
		if err := tx.Model(&match).Update("winner_id", match.P2ID).Error; err != nil {
			return err
		}
		if err := m.HandleByeByeMoving(match, match.P2ID, match.P1ID, tx); err != nil {
			return err
		}
	}
	if match.P2ID == math.MaxUint32 && match.P1ID != 0 && match.P1ID != math.MaxUint32 {
		if err := tx.Model(&match).Update("winner_id", match.P1ID).Error; err != nil {
			return err
		}
		if err := m.HandleByeByeMoving(match, match.P1ID, match.P2ID, tx); err != nil {
			return err
		}
	}
	//if both players are byebye-s
	if match.P1ID == math.MaxUint32 && match.P2ID == math.MaxUint32 {
		if err := tx.Model(&match).Update("winner_id", match.P1ID).Error; err != nil {
			return err
		}
		if err := m.HandleByeByeMoving(match, match.P1ID, match.P2ID, tx); err != nil {
			return err
		}
	}

	return nil
}

func (m *MatchRepository) HandleByeByeMoving(match model.Match, winner, loser uint, tx *gorm.DB) error {
	var nextWinnerMatch model.Match
	var nextLoserMatch model.Match
	//if final match then nexctwinnerminer is 0 do something idk what yet same with losermatch:W

	if err := tx.Model(&model.Match{}).Where("id = ?", match.WinnerNextMatch).Scan(&nextWinnerMatch).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.Match{}).Where("id = ?", match.LoserNextMatch).Scan(&nextLoserMatch).Error; err != nil {
		return err
	}

	if nextWinnerMatch.ID == 0 || nextLoserMatch.ID == 0 {
		return nil
	}

	if nextWinnerMatch.P1ID == 0 {
		//move player to p1id
		if err := tx.Model(&nextWinnerMatch).Update("p1_id", winner).Error; err != nil {
			return err
		}
	} else if nextWinnerMatch.P2ID == 0 {
		//move player to p2id
		if err := tx.Model(&nextWinnerMatch).Update("p2_id", winner).Error; err != nil {
			return err
		}
	}

	if nextLoserMatch.P1ID == 0 {
		//move player to p1id
		if err := tx.Model(&nextLoserMatch).Update("p1_id", loser).Error; err != nil {
			return err
		}
	} else if nextLoserMatch.P2ID == 0 {
		//move player to p2id
		if err := tx.Model(&nextLoserMatch).Update("p2_id", loser).Error; err != nil {
			return err
		}
	}

	//to stop recursion if we have arrived at the placement match
	if match.WinnerNextMatch == 0 || match.LoserNextMatch == 0 {
		return nil
	}

	// //recursivly check the next byebye-s also
	if err := m.CheckByeBye(nextLoserMatch, tx); err != nil {
		return err
	}
	if err := m.CheckByeBye(nextWinnerMatch, tx); err != nil {
		return err
	}
	return nil
}

//Newly-added stuff

func (m *MatchRepository) GetAllWinnerMatches(tournament *model.Tournament, tx *gorm.DB) ([]model.Match, error) {
	var output []model.Match
	if err := tx.Model(&model.Match{}).Where("tournament_id = ? AND type = 'winner'", tournament.ID).Find(&output).Error; err != nil {
		return nil, err
	}
	return output, nil
}

func (m *MatchRepository) GetAllLoserMatches(tournament *model.Tournament, tx *gorm.DB) ([]model.Match, error) {
	var output []model.Match
	if err := tx.Model(&model.Match{}).Where("tournament_id = ? AND type = 'loser'", tournament.ID).Find(&output).Error; err != nil {
		return nil, err
	}
	return output, nil
}

func (m *MatchRepository) GetAllBracketMatches(tournament *model.Tournament, from, to int, tx *gorm.DB) ([]model.Match, error) {
	var output []model.Match
	query := `
		tournament_id = ? 
		AND type = 'bracket' 
		AND bracket != '' 
		AND CAST(SPLIT_PART(bracket, '-', 1) AS INTEGER) = ? 
		AND CAST(SPLIT_PART(bracket, '-', 2) AS INTEGER) = ?`

	if err := tx.Model(&model.Match{}).
		Where(query, tournament.ID, from, to).
		Find(&output).Error; err != nil {
		return nil, err
	}
	return output, nil
}

func (m *MatchRepository) GetAllBracketMatchesWeak(tournament *model.Tournament, from, to int, tx *gorm.DB) ([]model.Match, error) {
	var output []model.Match
	query := `
	tournament_id = ?
	AND type = 'bracket'
	AND bracket != ''
	AND CAST(SPLIT_PART(bracket, '-', 1) AS INTEGER) >= ?
	AND CAST(SPLIT_PART(bracket, '-', 2) AS INTEGER) <= ?`

	if err := tx.Model(&model.Match{}).
		Where(query, tournament.ID, from, to).
		Order("CAST(SPLIT_PART(bracket, '-', 1) AS INTEGER) ASC, CAST(SPLIT_PART(bracket, '-', 2) AS INTEGER) DESC").
		Order("current_round ASC, identifier ASC").
		Find(&output).Error; err != nil {
		return nil, err
	}
	return output, nil
}

func (m *MatchRepository) GetTeamVSTeamMatch(tournament_id, team_1, team_2 uint, tx *gorm.DB) (*model.Match, error) {
	var match model.Match
	if err := tx.Model(&model.Match{}).Where("tournament_id = ? AND p1_id = ? AND p2_id = ?", tournament_id, team_1, team_2).Find(&match).Error; err != nil {
		return nil, err
	}
	return &match, nil
}

func (m *MatchRepository) UpdateForfeithMatch(match_id, winner_id uint, isForfeitMatch bool) error {
	//start transaction
	tx := m.store.Db.Begin()
	// resetting the fortfeit match
	if winner_id == 0 {
		// reset all the sets
		if err := m.store.Set().ResetMatchSets(match_id, tx); err != nil {
			return err
		}
	} else {
		//check if there is already winner id
		var match model.Match
		if err := tx.Model(&model.Match{}).Where("id = ?", match_id).Find(&match).Error; err != nil {
			return err
		}

		//update hte previous sets
		scores := make(map[int][2]int, 3)
		for setNumber := range 3 {
			if match.P1ID == winner_id {
				scores[setNumber+1] = [2]int{11, 0}
			} else {
				scores[setNumber+1] = [2]int{0, 11}
			}
		}
		if err := m.store.Set().UpdateMatchSets(match_id, scores, tx); err != nil {
			return err
		}

	}
	if err := tx.Model(&model.Match{}).Where("id = ?", match_id).Update("winner_id", winner_id).Update("forfeit_match", isForfeitMatch).Error; err != nil {
		return err
	}
	return tx.Commit().Error
}
