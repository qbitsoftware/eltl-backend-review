package psqlstore

import (
	"errors"
	"fmt"
	"table-tennis/internal/model"
)

type SessionRepository struct {
	store *Store
}

func (s *SessionRepository) Create(session *model.Session) (*model.Session, error) {
	if err := s.store.Db.Create(session).Error; err != nil {
		return nil, err
	}

	return session, nil
}

func (s *SessionRepository) CheckByUserId(userID uint) (*model.Session, error) {
	var session model.Session

	if err := s.store.Db.Where("user_id = ?", userID).First(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *SessionRepository) Update(accessID string, session model.Session) (*model.Session, error) {
	// Use GORM's Model, Where, and Updates methods

	fmt.Printf("%+v\n", session)
	if err := s.store.Db.Model(&model.Session{}).
		Where("acess_id = ?", accessID).
		Updates(map[string]interface{}{
			"acess_id":   session.AcessID,
			"user_id":    session.UserID,
			"created_at": session.CreatedAT,
		}).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *SessionRepository) Delete(accessID string) error {
	// Use GORM's Where and Delete methods
	if err := s.store.Db.Where("acess_id = ?", accessID).Delete(&model.Session{}).Error; err != nil {
		return err
	}

	// Check if any rows were affected
	if s.store.Db.RowsAffected == 0 {
		return errors.New("no rows affected")
	}

	return nil
}

func (s *SessionRepository) Check(accessID string) (*model.Session, error) {
	var session model.Session

	// Use GORM's Where and First methods
	if err := s.store.Db.Where("acess_id = ?", accessID).First(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

func (s *SessionRepository) DeleteByUser(userID uint) error {
	// Use GORM's Where and Delete methods
	if err := s.store.Db.Where("user_id = ?", userID).Delete(&model.Session{}).Error; err != nil {
		return err
	}

	return nil
}
