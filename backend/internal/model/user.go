package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	BirthDate       time.Time `json:"birth_date"`
	ClubID          string    `json:"club_id"`
	Email           string    `json:"email"`
	Password        string    `json:"-"`
	Sex             string    `json:"sex"`
	RatingPoints    int       `json:"rating_points"`
	PlacementPoints int       `json:"placement_points"`
	WeightPoints    int       `json:"weight_points"`
	EltlID          int       `json:"eltl_id"`
	HasRating       bool      `json:"has_rating"` //Valis mangija ---> kui on rating, ss on eesti, kui ei, ss on valis mangija
	Confirmation    string    `json:"confirmation"`
	Nationality     string    `json:"nationality"`
	PlacingOrder    int       `json:"placing_order"`
	Image_url string `json:"img_url"`
	Hv              int       `gorm:"-"`
	Hk              int       `gorm:"-"`
}

type PlayerWithTeam struct {
	TeamID       uint
	PlayerID     uint
	FirstName    string
	LastName     string
	Confirmation string
	HasRating    bool
}

func (u *User) Sanitize() {
	u.Password = ""
}

func (u *User) ComparePassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}

func EncryptString(str string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

//----------HOOKS------------//

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if len(u.Password) > 0 {
		enc, err := EncryptString(u.Password)
		if err != nil {
			return err
		}
		u.Password = enc
	}
	return nil
}

func (u *User) AfterFind(tx *gorm.DB) error {
	// u.Sanitize()
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed() {
		tx.Statement.SetColumn("UpdatedAt", time.Now())
	}
	return nil
}
