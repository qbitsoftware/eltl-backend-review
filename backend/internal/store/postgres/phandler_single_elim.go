package psqlstore

import (
	"fmt"
	"math"
	"table-tennis/internal/model"

	"gorm.io/gorm"
)

type PHandler_Single struct {
	store *Store
}

func CreatePHandlerSingle(database *Store) *PHandler_Single {
	return &PHandler_Single{
		store: database,
	}
}

func (p *PHandler_Single) MovePlayers(match *model.Match, tx *gorm.DB) error {
	winner, _, fm, err := p.store.Match().FindWinner(*match, tx)
	if err != nil {
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

	if fm {
		finished, err := p.store.Tournament().CheckIfFinised(match.TournamentID, tx)
		if err != nil {
			return err
		}
		if finished {
			//change tournament state
			if err := tx.Model(&model.Tournament{}).Where("id = ?", match.TournamentID).Update("state", "finished").Error; err != nil {
				return err
			}
			// calculate tournament ratings COMING SOON....
			fmt.Println("Calculating ratings...")
		}
		return nil
	}
	if err := p.NextMatch(*match, *winner, tx); err != nil {
		return err
	}

	return nil
}

func (p *PHandler_Single) NextMatch(match model.Match, winner uint, tx *gorm.DB) error {
	var nextWinnerMatch model.Match

	if err := tx.Model(&model.Match{}).Where("id = ?", match.WinnerNextMatch).Scan(&nextWinnerMatch).Error; err != nil {
		return err
	}

	if nextWinnerMatch.ID == 0 {
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

	//stop recursion if placement match
	if match.WinnerNextMatch == 0 {
		return nil
	}

	//check recursively byebyes
	if err := p.CheckByeBye(nextWinnerMatch, tx); err != nil {
		return err
	}

	return nil
}

func (p *PHandler_Single) CheckByeBye(match model.Match, tx *gorm.DB) error {
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
		if err := p.NextMatch(match, match.P2ID, tx); err != nil {
			return err
		}
	}
	if match.P2ID == math.MaxUint32 && match.P1ID != 0 && match.P1ID != math.MaxUint32 {
		if err := tx.Model(&match).Update("winner_id", match.P1ID).Error; err != nil {
			return err
		}
		if err := p.NextMatch(match, match.P1ID, tx); err != nil {
			return err
		}
	}
	//if both players are byebye-s
	if match.P1ID == math.MaxUint32 && match.P2ID == math.MaxUint32 {
		if err := tx.Model(&match).Update("winner_id", match.P1ID).Error; err != nil {
			return err
		}
		if err := p.NextMatch(match, match.P1ID, tx); err != nil {
			return err
		}
	}
	return nil
}
