package model

import "gorm.io/gorm"

type Club struct {
	gorm.Model
	Name        string
	ContactName string
	Email       *string
	PhoneNumber *string
	PostAddress *string
	Url         *string
}
