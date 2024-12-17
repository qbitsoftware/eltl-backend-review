package psqlstore

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"table-tennis/internal/model"
	"time"

	"github.com/google/uuid"
)

type UserRepository struct {
	store *Store
}

func (u *UserRepository) Create(user model.User) error {
	if result := u.store.Db.Create(&user); result.Error != nil {
		return result.Error
	}
	return nil
}

// func (u *UserRepository) GetByLogin(email string) (*model.User, error) {
// 	var user model.User
// 	if result := u.store.Db.Model(&model.User{}).Where("email = ?", email).First(&user).Error; result != nil {
// 		return nil, result
// 	}
// 	return &user, nil
// }

func (u *UserRepository) Get(id uint) (*model.User, error) {
	var user model.User
	if result := u.store.Db.First(&user, "id =?", id); result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (u *UserRepository) GetAll() ([]model.User, error) {
	var users []model.User
	if err := u.store.Db.Model(&model.User{}).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (u *UserRepository) GetPlayers(mId uint) ([]model.User, error) {
	var users []model.User
	var p1, p2 model.User

	// Get player 1
	if result := u.store.Db.Model(&model.User{}).
		Joins("left join matches on matches.p1_id = users.id").
		Where("matches.id = ?", mId).
		Scan(&p1).Error; result != nil {
		return nil, result
	}

	// Get player 2
	if result := u.store.Db.Model(&model.User{}).
		Joins("left join matches on matches.p2_id = users.id").
		Where("matches.id = ?", mId).
		Scan(&p2).Error; result != nil {
		return nil, result
	}

	users = append(users, p1, p2)

	return users, nil
}

func (u *UserRepository) GetMatches(uId uint) ([]model.Match, error) {
	var matches []model.Match
	if err := u.store.Db.Model(&model.Match{}).Where("(p1_id = ? OR p2_id = ?) AND winner_id != 0", uId, uId).Find(&matches).Error; err != nil {
		return nil, err
	}
	fmt.Println(matches)
	return matches, nil
}

func (u *UserRepository) CreateBlog(blog model.Blog) error {
	if err := u.store.Db.Save(&blog).Error; err != nil {
		return err
	}
	return nil
}

func (u *UserRepository) GetTournamentBlog(tournament_id uint) (*model.Blog, error) {
	var output model.Blog
	if result := u.store.Db.Model(&model.Blog{}).
		Joins("left join users on blogs.author_id = users.id").
		Where("blogs.tournament_id = ?", tournament_id).
		First(&output).Error; result != nil {
		return nil, result
	}
	return &output, nil
}

func (u *UserRepository) ScrapeUsers() error {
	url := "https://www.lauatennis.ee/app_partner/app_reiting_xml.php"

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Create a variable of the struct type to store the parsed data
	var personReitings PersonXReitings

	// Unmarshal the XML data into the struct
	err = xml.Unmarshal(body, &personReitings)
	if err != nil {
		return err
	}

	for _, person := range personReitings.PersonXReitings {
		birthDate, err := time.Parse("2006-01-02", person.BirthDate)
		if err != nil {
			return err
		}

		if err := u.store.Db.Create(&model.User{
			FirstName:       person.FirstName,
			LastName:        person.FamName,
			BirthDate:       birthDate,
			Email:           "placeholder@gmail.com",
			Password:        uuid.New().String(),
			ClubID:          person.ClbName,
			RatingPoints:    person.RatePoints,
			PlacementPoints: person.RatePlPnts,
			WeightPoints:    person.RateWeight,
			EltlID:          person.PersonID,
			Sex:             person.Sex,
			HasRating:       true,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

type PersonXReitings struct {
	XMLName         xml.Name         `xml:"person_x_reitings"`
	PersonXReitings []PersonXReiting `xml:"person_x_reiting"`
}

type PersonXReiting struct {
	PersonID   int    `xml:"personid"`
	FamName    string `xml:"famname"`
	FirstName  string `xml:"firstname"`
	Sex        string `xml:"sex"`
	BirthDate  string `xml:"birthdate"`
	RateDate   string `xml:"ratedate"`
	RateOrder  int    `xml:"rateorder"`
	RatePlPnts int    `xml:"rateplpnts"`
	RatePoints int    `xml:"ratepoints"`
	RateWeight int    `xml:"rateweight"`
	RateWeiLow string `xml:"rateweilow"`
	ClbName    string `xml:"clbname"`
}
