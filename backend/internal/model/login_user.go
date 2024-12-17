package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginUser struct {
	gorm.Model
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"-"`
	Email     string `json:"email"`
	Role      uint   `json:"role"`
}

func (u *LoginUser) Sanitize() {
	u.Password = ""
}

func (u *LoginUser) ComparePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

//----------HOOKS------------//

func (u *LoginUser) BeforeCreate(tx *gorm.DB) error {
	if len(u.Password) > 0 {
		enc, err := EncryptString(u.Password)
		if err != nil {
			return err
		}
		u.Password = enc
	}
	return nil
}

func (u *LoginUser) AfterFind(tx *gorm.DB) error {
	// u.Sanitize()
	return nil
}

func (u *LoginUser) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed() {
		tx.Statement.SetColumn("UpdatedAt", time.Now())
	}
	return nil
}
