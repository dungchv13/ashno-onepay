package repository

import (
	"ashno-onepay/internal/model"
	"gorm.io/gorm"
	"sync"
)

type RegistrationOptionRepository interface {
	Find(req model.RegistrationOptionFilter) (*model.RegistrationOption, error)
}

type registrationOptionRepository struct {
	db *gorm.DB
}

func (r registrationOptionRepository) Find(req model.RegistrationOptionFilter) (*model.RegistrationOption, error) {
	var option model.RegistrationOption

	query := r.db.Model(&model.RegistrationOption{})

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Subtype != "" {
		query = query.Where("subtype = ?", req.Subtype)
	}

	err := query.First(&option).Error
	return &option, err
}

var registrationOptionRepositoryInstance *registrationOptionRepository
var registrationOptionRepositoryOnce sync.Once

func GetRegistrationOptionRepositoryInstance(db *gorm.DB) RegistrationOptionRepository {
	registrationOptionRepositoryOnce.Do(func() {
		registrationOptionRepositoryInstance = &registrationOptionRepository{
			db: db,
		}
	})
	return registrationOptionRepositoryInstance
}
