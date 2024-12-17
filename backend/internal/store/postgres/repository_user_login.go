package psqlstore

import "table-tennis/internal/model"

type UserLoginRepository struct {
	store *Store
}

func (u *UserLoginRepository) GetByLogin(email string) (*model.LoginUser, error) {
	var user model.LoginUser
	if result := u.store.Db.Model(&model.LoginUser{}).Where("email = ?", email).First(&user).Error; result != nil {
		return nil, result
	}
	return &user, nil
}

func (u *UserLoginRepository) Get(id uint) (*model.LoginUser, error) {
	var user model.LoginUser
	if result := u.store.Db.First(&user, "id =?", id); result.Error != nil {
		return nil, result.Error
	}

	return &user, nil
}

func (u *UserLoginRepository) Create(user model.LoginUser) error {
	if result := u.store.Db.Create(&user); result.Error != nil {
		return result.Error
	}
	return nil
}
