package service

import (
	"ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"ashno-onepay/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"sync"
)

type UserService interface {
	UserLogin(email, password string) (model.User, error)
	UserRegister(user model.User) error
}

type userService struct {
	userRepo repository.UserRepository
}

func (us userService) UserRegister(user model.User) error {
	_, err := us.userRepo.GetUserByEmail(user.Email)
	if err == nil {
		return errors.ErrInvalidPassword.Reform("email registered")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	user.Password, err = HashPassword(user.Password)
	if err != nil {
		return err
	}
	return us.userRepo.SaveUserInfo(user)
}

func (us userService) UserLogin(email, password string) (model.User, error) {
	user, err := us.userRepo.GetUserByEmail(email)
	if err != nil {
		return model.User{}, err
	}
	if !CheckPasswordHash(password, user.Password) {
		return model.User{}, errors.ErrInvalidPassword
	}
	return user, nil
}

// Hash password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Compare hashed password with plain text
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

var userServiceInstance UserService
var userServiceOnce sync.Once

func GetUserServiceInstance(
	userRepo repository.UserRepository,
) UserService {
	userServiceOnce.Do(func() {
		userServiceInstance = NewUserService(
			userRepo,
		)
	})
	return userServiceInstance
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}
