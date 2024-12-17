package psqlstore

import (
	"errors"
	"fmt"
	"math"
	"table-tennis/internal/model"
	"table-tennis/internal/store/utils"

	"gorm.io/gorm"
)

type Generator_Double struct {
	Generator_Base
	phandler *PHandler_Double
}

func CreateGeneratorDouble(database *Store) *Generator_Double {
	return &Generator_Double{
		Generator_Base: Generator_Base{
			store: database,
		},
		phandler: CreatePHandlerDouble(database),
	}
}

func (d *Generator_Double) CreateMatches(tournament *model.Tournament, tx *gorm.DB) error {
	fmt.Println("DOUBLE ELIMINATION: Creating matches...")
	participants, err := d.store.Tournament().GetParticipants(tournament.ID)
	if err != nil {
		return err
	}

	if tournament.State != "started" {
		return fmt.Errorf("tournament is not in valid state, want - started, have - %v", tournament.State)
	}

	w, l := utils.CalculateMatchesPerRoundDouble(len(participants))
	//winners bracket
	for i := 1; i <= len(w); i++ {
		for j := 1; j <= w[i]; j++ {
			var bracket string
			if i == len(w) && j == w[i] {
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
	//losers bracket
	for i := 1; i <= len(l); i++ {
		for j := 1; j <= l[i]; j++ {
			var bracket string
			if i == len(l) && j == l[i] {
				bracket = "3-4"
			}
			m := model.Match{
				TournamentID: tournament.ID,
				Type:         "loser",
				CurrentRound: i,
				Identifier:   j,
				Bracket:      bracket,
			}
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
	}
	if err := d.UpdateTournamentState(tournament, tx); err != nil {
		return err
	}
	return nil
}

func (d *Generator_Double) CreateBrackets(tournament *model.Tournament, tx *gorm.DB) error {
	fmt.Println("DOUBLE ELIMINATION: Creating brackets....")
	if err := d.DE_CreateBrackets(tournament, tx); err != nil {
		return err
	}
	return nil
}

func (d *Generator_Double) LinkMatches(tournament *model.Tournament, tx *gorm.DB) error {
	fmt.Println("DOUBLE ELIMINATION: Linking matches...")
	if err := d.LinkWinners(tournament, tx); err != nil {
		return err
	}
	if err := d.LinkLosers(tournament, tx); err != nil {
		return err
	}
	if err := d.DE_LinkBrackets(d.store, tournament, tx); err != nil {
		return err
	}
	return nil
}

func (d *Generator_Double) AssignPlayers(tournament *model.Tournament, tx *gorm.DB) error {
	matches, err := d.store.Match().GetRoundMatches(1, tournament.ID, tx, "winner")
	if err != nil {
		return err
	}

	if len(matches) <= 0 {
		return errors.New("no matches were found")
	}

	participants, err := d.store.Tournament().GetParticipants(tournament.ID)
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
			//automatically send player to the next round
			if err := d.phandler.CheckByeBye(matches[i], tx); err != nil {
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
			//automatically send player to the next round
			if err := d.phandler.CheckByeBye(matches[i], tx); err != nil {
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

func (d *Generator_Double) LinkWinners(tournament *model.Tournament, tx *gorm.DB) error {
	winnerMatches, err := d.store.Match().GetAllWinnerMatches(tournament, tx)
	if err != nil {
		return err
	}
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
		var loserMatch model.Match
		var winnerMatch model.Match
		//next winner match
		nextWinnerMatchRound := match.CurrentRound + 1
		nextWinnerMatchIdentifier := math.Ceil(float64(match.Identifier) / 2)
		//next loser match
		var nextLoserMatchRound int
		var nextLoserMatchIdentifier float64
		if match.CurrentRound == 1 {
			nextLoserMatchIdentifier = math.Ceil(float64(match.Identifier) / 2)
			nextLoserMatchRound = match.CurrentRound
		} else {
			nextLoserMatchIdentifier = float64(match.Identifier)
			nextLoserMatchRound = match.CurrentRound + (match.CurrentRound - 2)
		}

		//update the match (find the right match id-s)
		if err := tx.Model(&model.Match{}).Where("current_round = ? AND identifier = ? AND tournament_id = ? AND type = 'loser'", nextLoserMatchRound, nextLoserMatchIdentifier, match.TournamentID).First(&loserMatch).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.Match{}).Where("current_round = ? AND identifier = ? AND tournament_id = ?", nextWinnerMatchRound, nextWinnerMatchIdentifier, match.TournamentID).First(&winnerMatch).Error; err != nil {
			return err
		}
		if err := tx.Model(&match).Update("winner_next_match", winnerMatch.ID).Error; err != nil {
			return err
		}
		if err := tx.Model(&match).Update("loser_next_match", loserMatch.ID).Error; err != nil {
			return err
		}
	}
	return nil
}

func (d *Generator_Double) LinkLosers(tournament *model.Tournament, tx *gorm.DB) error {
	loserMatches, err := d.store.Match().GetAllLoserMatches(tournament, tx)
	if err != nil {
		return err
	}

	brackets, err := d.store.Tournament().GetBrackets(tournament, tx)
	if err != nil {
		return err
	}

	for _, match := range loserMatches {
		bracket, err := utils.ParseBracketString(match.Bracket)
		if err != nil {
			return err
		}
		//final round don't add any next round
		if bracket[1]-bracket[0] == 1 {
			continue
		}

		matchCurrentBracket, err := utils.ParseBracketString(brackets[match.CurrentRound])
		if err != nil {
			return err
		}
		matchBrackets, err := d.store.Match().GetAllBracketMatches(tournament, matchCurrentBracket[0], matchCurrentBracket[1], tx)
		if err != nil {
			return err
		}

		nextWinnerMatchRound := match.CurrentRound + 1
		nextWinnerMatchIdentifier := math.Ceil(float64(match.Identifier) / 2)
		//next loser match
		nextLoserMatchRound := 1
		nextLoserMatchIdentifier := math.Ceil(float64(match.Identifier) / 2)

		if (match.CurrentRound+1)%2 == 0 {
			nextWinnerMatchIdentifier = float64(match.Identifier)
		}

		var winnerMatch model.Match
		if err := tx.Model(&model.Match{}).Where("current_round = ? AND identifier = ? AND tournament_id = ? AND type = 'loser'", nextWinnerMatchRound, nextWinnerMatchIdentifier, match.TournamentID).First(&winnerMatch).Error; err != nil {
			return err
		}
		if err := tx.Model(&match).Update("winner_next_match", winnerMatch.ID).Error; err != nil {
			return err
		}

		//if we have 1 match, it means its placement match awnyways we don't need to filter
		if len(matchBrackets) == 1 {
			if err := tx.Model(&match).Update("loser_next_match", matchBrackets[0].ID).Error; err != nil {
				return err
			}
		} else {
			for _, mB := range matchBrackets {
				if mB.CurrentRound == nextLoserMatchRound && mB.Identifier == int(nextLoserMatchIdentifier) {
					if err := tx.Model(&match).Update("loser_next_match", mB.ID).Error; err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
