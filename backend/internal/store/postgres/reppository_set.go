package psqlstore

import (
	"errors"
	"table-tennis/internal/model"

	"gorm.io/gorm"
)

type SetRepository struct {
	store *Store
}

func (s *SetRepository) Create(set model.Set) error {
	if err := s.store.Db.Create(&set).Error; err != nil {
		return err
	}
	return nil
}

func (s *SetRepository) Update(set model.Set) error {
	var existingSet model.Set
	if err := s.store.Db.Where("id = ? AND match_id = ?", set.ID, set.MatchID).First(&existingSet).Error; err != nil {
		return err
	}

	if !existingSet.IsEqual(set) {
		if result := s.store.Db.Model(&model.Set{}).
			Where("id = ? AND match_id = ?", set.ID, set.MatchID).
			Updates(map[string]interface{}{"team1_score": set.Team1Score, "team2_score": set.Team2Score, "set_number": set.SetNumber}); result.Error != nil {
			return result.Error
		}
		return nil
	} else {
		return errors.New("no changes detected")
	}
}

func (s *SetRepository) ResetMatchSets(matchID uint, tx *gorm.DB) error {
	if err := tx.Model(&model.Set{}).Where("match_id = ?", matchID).
		Updates(map[string]interface{}{
			"team1_score": 0,
			"team2_score": 0,
		}).Error; err != nil {
		return err
	}
	return nil
}

func (s *SetRepository) UpdateMatchSets(matchID uint, scores map[int][2]int, tx *gorm.DB) error {
	var sets []model.Set
	if err := tx.Where("match_id = ?", matchID).Find(&sets).Error; err != nil {
		return err
	}

	for i, set := range sets {
		if score, ok := scores[set.SetNumber]; ok {
			sets[i].Team1Score = score[0]
			sets[i].Team2Score = score[1]
		}
	}

	if err := tx.Save(&sets).Error; err != nil {
		return err
	}

	return nil
}
