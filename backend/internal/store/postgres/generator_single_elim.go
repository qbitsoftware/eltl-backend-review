package psqlstore

import (
	"errors"
	"fmt"
	"math"
	"table-tennis/internal/model"
	"table-tennis/internal/store/utils"

	"gorm.io/gorm"
)

type Generator_Single struct {
	Generator_Base
	phandler *PHandler_Single
}

func CreateGeneratorSingle(database *Store) *Generator_Single {
	return &Generator_Single{
		phandler: CreatePHandlerSingle(database),
		Generator_Base: Generator_Base{
			store: database,
		},
	}
}

func (s *Generator_Single) CreateMatches(tournament *model.Tournament, tx *gorm.DB) error {
	fmt.Println("SINGLE ELIMINATION: Creating matches...")
	participants, err := s.store.Tournament().GetParticipants(tournament.ID)
	if err != nil {
		return err
	}

	if tournament.State != "started" {
		return fmt.Errorf("tournament is not in valid state, want - started, have - %v", tournament.State)
	}

	m := utils.CalculateMatchesPerRoundSingle(len(participants))
	for i := 1; i <= len(m); i++ {
		for j := 1; j <= m[i]; j++ {
			var bracket string
			if i == len(m) && j == m[i] {
				bracket = "1-2"
			}
			m := model.Match{
				TournamentID: tournament.ID,
				Type:         "winner",
				CurrentRound: i,
				Identifier:   j,
				Bracket:      bracket,
			}
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
	}
	if err := s.UpdateTournamentState(tournament, tx); err != nil {
		return err
	}
	return nil
}

func (s *Generator_Single) CreateBrackets(tournament *model.Tournament, tx *gorm.DB) error {
	fmt.Println("SINGLE ELIMINATION: Skipping brackets...")
	return nil
}

func (s *Generator_Single) LinkMatches(tournament *model.Tournament, tx *gorm.DB) error {
	fmt.Println("SINGLE ELIMINATION: Linking matches...")
	winnerMatches, err := s.store.Match().GetAllWinnerMatches(tournament, tx)
	if err != nil {
		return err
	}
	// for _, m := range winnerMatches {
	// 	fmt.Println(m)
	// }
	for _, match := range winnerMatches {
		//check if its the final round
		bracket, err := utils.ParseBracketString(match.Bracket)
		if err != nil {
			return err
		}
		//final round don't add any next round
		if bracket[1]-bracket[0] == 1 {
			continue
		}
		var winnerMatch model.Match
		//next winner match
		nextWinnerMatchRound := match.CurrentRound + 1
		nextWinnerMatchIdentifier := math.Ceil(float64(match.Identifier) / 2)

		//update the match (find the right match id-s)
		if err := tx.Model(&model.Match{}).Where("current_round = ? AND identifier = ? AND tournament_id = ?", nextWinnerMatchRound, nextWinnerMatchIdentifier, match.TournamentID).First(&winnerMatch).Error; err != nil {
			return err
		}
		if err := tx.Model(&match).Update("winner_next_match", winnerMatch.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Generator_Single) AssignPlayers(tournament *model.Tournament, tx *gorm.DB) error {
	fmt.Println("SINGLE ELIMINATION: Assigning players...")
	matches, err := s.store.Match().GetRoundMatches(1, tournament.ID, tx, "winner")
	if err != nil {
		return err
	}

	if len(matches) <= 0 {
		return errors.New("no matches were found")
	}

	participants, err := s.store.Tournament().GetParticipants(tournament.ID)
	if err != nil {
		return err
	}

	//get players / teams
	numberTeams := len(matches) * 2
	limit := int(math.Log2(float64(numberTeams)) + 1)
	players := utils.Branch(1, 1, limit)

	for i := 0; i < len(matches); i++ {
		//player 1
		if players[i][0]-1 >= len(participants) {
			if err := tx.Model(&model.Match{}).Where("id = ?", matches[i].ID).Update("p1_id", math.MaxUint32).Error; err != nil {
				return err
			}
			//check byebye movement
			if err := s.phandler.CheckByeBye(matches[i], tx); err != nil {
				return err
			}
		} else {
			if err := tx.Model(&model.Match{}).Where("id = ?", matches[i].ID).Update("p1_id", participants[players[i][0]-1].ID).Error; err != nil {
				return err
			}
		}
		//player 2
		if players[i][1]-1 >= len(participants) {
			if err := tx.Model(&model.Match{}).Where("id = ?", matches[i].ID).Update("p2_id", math.MaxUint32).Error; err != nil {
				return err
			}
			//check byebye movement
			if err := s.phandler.CheckByeBye(matches[i], tx); err != nil {
				return err
			}
		} else {
			if err := tx.Model(&model.Match{}).Where("id = ?", matches[i].ID).Update("p2_id", participants[players[i][1]-1].ID).Error; err != nil {
				return err
			}
		}
	}

	// update tournament state
	if result := tx.Model(&model.Tournament{}).Where("id = ?", tournament.ID).Update("state", "players_assigned"); result.Error != nil {
		return result.Error
	}
	return nil
}
