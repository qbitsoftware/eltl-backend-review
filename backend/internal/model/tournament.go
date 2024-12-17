package model

import (
	"time"

	"gorm.io/gorm"
)

type Tournament struct {
	gorm.Model
	Name      string     `json:"name"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
	Type      string     `json:"type"`
	State     string     `json:"state"`
	Private   bool       `json:"private"`
}

func (t *Tournament) IsEqual(other Tournament) bool {
	if t.Name == other.Name &&
		t.StartDate.Equal(*other.StartDate) &&
		t.EndDate.Equal(*other.EndDate) &&
		t.Type == other.Type &&
		t.Private == other.Private &&
		t.State == other.State {
		return true
	} else {
		return false
	}
}

//----------HOOKS------------//

func (t *Tournament) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed() {
		tx.Statement.SetColumn("UpdatedAt", time.Now())
	}

	return nil
}
