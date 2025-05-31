package repository

import (
	"ashno-onepay/internal/model"
	"gorm.io/gorm"
	"sync"
)

type UserRepository interface {
	GetUserByEmail(email string) (model.User, error)
	SaveUserInfo(user model.User) error
}

type userRepository struct {
	db *gorm.DB
}

func (u userRepository) SaveUserInfo(user model.User) error {
	result := u.db.Create(&user)
	return result.Error
}

func (u userRepository) GetUserByEmail(email string) (model.User, error) {
	var user model.User
	result := u.db.First(&user, model.User{Email: email})
	return user, result.Error
}

var usersRepositoryInstance *userRepository
var usersRepositoryOnce sync.Once

func GetUserRepositoryInstance(db *gorm.DB) UserRepository {
	usersRepositoryOnce.Do(func() {
		usersRepositoryInstance = &userRepository{
			db: db,
		}
	})
	return usersRepositoryInstance
}
