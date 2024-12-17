package psqlstore

import (
	"fmt"
	"math"
	"table-tennis/internal/model"
	"table-tennis/internal/store/utils"

	"gorm.io/gorm"
)

type Generator_Base struct {
	store *Store
}

func (b *Generator_Base) UpdateTournamentState(tournament *model.Tournament, tx *gorm.DB) error {
	var initialState string
	if err := tx.Model(&model.Tournament{}).Select("state").Where("id = ?", tournament.ID).Row().Scan(&initialState); err != nil {
		return err
	}

	if err := tx.Model(&model.Tournament{}).Where("id = ?", tournament.ID).Update("state", "matches_created").Error; err != nil {
		return err
	}

	var updatedState string
	if err := tx.Model(&model.Tournament{}).Select("state").Where("id = ?", tournament.ID).Row().Scan(&updatedState); err != nil {
		return err
	}

	if updatedState != "matches_created" {
		return fmt.Errorf("tournament state was not updated correctly, expected - matches_created, got - %v", tournament.State)
	}

	return nil
}

func (b *Generator_Base) DE_CreateBrackets(tournament *model.Tournament, tx *gorm.DB) error {
	maxRounds, err := b.store.Match().GetLoserMaxRounds(tournament.ID, tx)
	if err != nil {
		return err
	}

	//first of all create bracket
	lastRoundParticipants, err := b.store.Tournament().GetParticipants(tournament.ID)
	if err != nil {
		return err
	}

	last_participants := utils.RoundUpToPowerOf2(len(lastRoundParticipants))
	for i := 1; i < maxRounds; i++ {
		participants, err := b.store.Match().GetLoserCurrentRoundParticipants(tournament.ID, i, tx)
		if err != nil {
			return err
		}

		until := last_participants - (participants/2 - 1)

		//if we have less than 4 participants or less (less than 2 matches then just change the current round for 1 for brackets bcz its just placement match)
		if err := b.store.Tournament().CreateBracketRecursive(tournament, until, last_participants, i, tx); err != nil {
			return err
		}

		last_participants = until - 1

	}
	return nil
}

func (b *Generator_Base) DE_LinkBrackets(s *Store, tournament *model.Tournament, tx *gorm.DB) error {
	brackets, err := s.Tournament().GetBrackets(tournament, tx)
	if err != nil {
		return err
	}
	for _, value := range brackets {
		matchCurrentBracket, err := utils.ParseBracketString(value)
		if err != nil {
			return err
		}

		matchesToLink, err := s.Match().GetAllBracketMatchesWeak(tournament, matchCurrentBracket[0], matchCurrentBracket[1], tx)
		if err != nil {
			return err
		}

		for _, ma := range matchesToLink {
			parsedBrackets, err := utils.ParseBracketString(ma.Bracket)
			if err != nil {
				return err
			}
			//placement match does not need next rounds
			if parsedBrackets[1]-parsedBrackets[0] == 1 {
				continue
			}

			nextMatchIdentifier := math.Ceil(float64(ma.Identifier) / 2)
			offSet := ((parsedBrackets[1] - parsedBrackets[0]) + 1) / 2
			nextWinnerBracket := fmt.Sprintf("%d-%d", parsedBrackets[0], parsedBrackets[1]-offSet)
			nextLoserBracket := fmt.Sprintf("%d-%d", parsedBrackets[0]+offSet, parsedBrackets[1])

			for _, nm := range matchesToLink {
				if nm.Bracket == nextWinnerBracket && nm.Identifier == int(nextMatchIdentifier) {
					if err := tx.Model(&ma).Update("winner_next_match", nm.ID).Error; err != nil {
						return err
					}
				} else if nm.Bracket == nextLoserBracket && nm.Identifier == int(nextMatchIdentifier) {
					if err := tx.Model(&ma).Update("loser_next_match", nm.ID).Error; err != nil {
						return err
					}
				}
				//if its final then the loser is on the same level as winners bracket but wtih identifier offset + 1
				if parsedBrackets[1]-parsedBrackets[0] == 3 {
					if nm.Bracket == nextLoserBracket && nm.Identifier == int(nextMatchIdentifier)+1 {
						if err := tx.Model(&ma).Update("loser_next_match", nm.ID).Error; err != nil {
							return err
						}
					}
				}
			}
		}
	}
	return nil
}
