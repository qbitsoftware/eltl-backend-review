package psqlstore

import (
	"fmt"
	"math"
	"table-tennis/internal/model"

	"gorm.io/gorm"
)

type PHandler_Double struct {
	store *Store
}

func CreatePHandlerDouble(database *Store) *PHandler_Double {
	return &PHandler_Double{
		store: database,
	}
}

func (p *PHandler_Double) MovePlayers(match *model.Match, tx *gorm.DB) error {
	fmt.Println("DOUBLE ELIMINATION: Moving players...")
	winner, loser, fm, err := p.store.Match().FindWinner(*match, tx)
	if err != nil {
		return err
	}
	//query the match
	if err := tx.Model(&model.Match{}).Where("id = ?", match.ID).Scan(&match).Error; err != nil {
		return err
	}

	//update current match winnerid
	if err := tx.Model(&match).Update("winner_id", winner).Error; err != nil {
		return err
	}
	//if placement match
	if fm {
		finished, err := p.store.Tournament().CheckIfFinised(match.TournamentID, tx)
		if err != nil {
			return err
		}
		if finished {
			if err := tx.Model(&model.Tournament{}).Where("id = ?", match.TournamentID).Update("state", "finished").Error; err != nil {
				return err
			}
			fmt.Println("Starting calculating...")
		}

		return nil
	}

	if err := p.NextMatch(*match, *winner, *loser, tx); err != nil {
		return err
	}

	return nil

}

func (p *PHandler_Double) CheckByeBye(match model.Match, tx *gorm.DB) error {
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
		if err := p.NextMatch(match, match.P2ID, match.P1ID, tx); err != nil {
			return err
		}
	}
	if match.P2ID == math.MaxUint32 && match.P1ID != 0 && match.P1ID != math.MaxUint32 {
		if err := tx.Model(&match).Update("winner_id", match.P1ID).Error; err != nil {
			return err
		}
		if err := p.NextMatch(match, match.P1ID, match.P2ID, tx); err != nil {
			return err
		}
	}
	//if both players are byebye-s
	if match.P1ID == math.MaxUint32 && match.P2ID == math.MaxUint32 {
		if err := tx.Model(&match).Update("winner_id", match.P1ID).Error; err != nil {
			return err
		}
		if err := p.NextMatch(match, match.P1ID, match.P2ID, tx); err != nil {
			return err
		}
	}
	return nil
}

func (p *PHandler_Double) NextMatch(match model.Match, winner, loser uint, tx *gorm.DB) error {
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
	if err := p.CheckByeBye(nextLoserMatch, tx); err != nil {
		return err
	}
	if err := p.CheckByeBye(nextWinnerMatch, tx); err != nil {
		return err
	}
	return nil
}

