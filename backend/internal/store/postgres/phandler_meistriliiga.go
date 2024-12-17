package psqlstore

import (
	"errors"
	"table-tennis/internal/model"

	"gorm.io/gorm"
)

type PHandler_Meistriliiga struct {
	store *Store
}

func CreatePHandlerMeistriliiga(database *Store) *PHandler_Meistriliiga {
	return &PHandler_Meistriliiga{
		store: database,
	}
}

func (p *PHandler_Meistriliiga) MovePlayers(match *model.Match, tx *gorm.DB) error {
	team_match, err := p.store.Club().GetTeamMatch(match.ID)
	if err != nil {
		return err
	}
	h_referee := match.HeadReferee
	t_referee := match.TableReferee
	if h_referee == "" || t_referee == "" {
		return errors.New("kohtunikud on kohustuslikud")
	}

	if err := p.store.Db.Find(&match, match.ID).Error; err != nil {
		return err
	}

	if match.WinnerID != 0 {
		return errors.New("võitja on juba selgunud")
	}
	//ASSINGMENT AGAIN
	match.HeadReferee = h_referee
	match.TableReferee = t_referee

	team_matches, err := p.store.Club().GetTeamMatches(team_match.TournamentID, team_match.ID)
	if err != nil {
		return err
	}
	p1_points := 0
	p2_points := 0
	for _, t_match := range team_matches {
		var p1points int
		var p2points int
		sets, err := p.store.Match().GetMatchSets(t_match.ID)
		if err != nil {
			return err
		}
		for _, set := range sets {
			if set.Team1Score > set.Team2Score {
				p1points++
			} else if set.Team2Score > set.Team1Score {
				p2points++
			}
		}
		if p1points > p2points && p1points >= 3 {
			p1_points++
		} else if p2points > p1points && p2points >= 3 {
			p2_points++
		}
	}

	if p1_points < 4 && p2_points < 4 {
		return errors.New("voistlus peab kaima, kes esimesena saab 4 punktimajandust")
	}

	if p1_points > p2_points {
		//p1 wins
		if err := tx.Model(&model.Match{}).Where("id = ?", match.ID).Updates(
			map[string]interface{}{
				"winner_id":     match.P1ID,
				"head_referee":  match.HeadReferee,
				"table_referee": match.TableReferee,
			},
		).Error; err != nil {
			return err
		}
	} else if p2_points > p1_points {
		//p2 wins
		if err := tx.Model(&model.Match{}).Where("id = ?", match.ID).Updates(
			map[string]interface{}{
				"winner_id":     match.P2ID,
				"head_referee":  match.HeadReferee,
				"table_referee": match.TableReferee,
			},
		).Error; err != nil {
			return err
		}
	} else {
		return errors.New("draw")
	}

	//if no matches are remaining generate the last ones
	// matches, err := p.store.Match().GetRemainingMeistriliigaMatches(match.TournamentID, tx)
	// if err != nil {
	// 	return err
	// }
	// if len(matches) == 0 {
	// 	//Generate regrouping matches
	// 	//First of all, reorded the table according to the points in the tournaments
	// 	all_teams_regrouped, err := p.store.Club().GetTournamentTeamsRegroupted(match.TournamentID, tx)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if err := CreateRegroupedMatches(reversePairs(MeistriLiigaMatches), match.TournamentID, all_teams_regrouped, tx); err != nil {
	// 		return err
	// 	}
	// }
	return nil
}

// func CreateRegroupedMatches(newMatches [7][4][2]int, tournament_id uint, all_teams_regrouped []model.Team, tx *gorm.DB) error {
// 	//query tournament
// 	var tournament model.Tournament
// 	if err := tx.Model(model.Tournament{}).Where("id = ?", tournament_id).Find(&tournament).Error; err != nil {
// 		return err
// 	}
// 	//it always starts from the third day
// 	gameDay := 3
// 	//we have already generated 7 rounds
// 	roundNumber := 8

// 	for i, round := range newMatches {
// 		if i == 1 || i == 4 {
// 			gameDay++
// 		}
// 		for index, match := range round {
// 			year, month, day := tournament.StartDate.Date()
// 			var start_date time.Time
// 			if roundNumber == 8 || roundNumber == 10 || roundNumber == 13 {
// 				start_date = time.Date(year, month, day+gameDay-1, 13, 0, 0, 0, tournament.StartDate.Location())
// 			} else if roundNumber == 9 || roundNumber == 12 {
// 				start_date = time.Date(year, month, day+gameDay-1, 10, 0, 0, 0, tournament.StartDate.Location())
// 			} else if roundNumber == 11 || roundNumber == 14 {
// 				start_date = time.Date(year, month, day+gameDay-1, 16, 0, 0, 0, tournament.StartDate.Location())
// 			}
// 			matchToCreate := model.Match{
// 				TournamentID: tournament_id,
// 				P1ID:         all_teams_regrouped[match[0]-1].ID,
// 				P2ID:         all_teams_regrouped[match[1]-1].ID,
// 				Type:         "voor",
// 				CurrentRound: roundNumber, //Vooru nr
// 				Identifier:   gameDay,     //Game day number
// 				StartDate:    start_date,
// 				Table:        index + 1,
// 			}
// 			if err := tx.Create(&matchToCreate).Error; err != nil {
// 				return err
// 			}
// 		}
// 		roundNumber++
// 	}
// 	return nil
// }

func reversePairs(matches [7][4][2]int) [7][4][2]int {
	for i := range matches {
		for j := range matches[i] {
			matches[i][j][0], matches[i][j][1] = matches[i][j][1], matches[i][j][0]
		}
	}
	return matches
}
