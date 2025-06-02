package repository

import (
	errs "ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"errors"
	"gorm.io/gorm"
	"sync"
)

type RegistrationRepository interface {
	Create(registration model.Registration) (*model.Registration, error)
	GetByEmail(email string) (*model.Registration, error)
	GetRegistration(ID string) (*model.Registration, error)
	UpdatePaymentStatus(ID, status string) error
	Remove(ID string) error
}

type registrationRepository struct {
	db *gorm.DB
}

func (r registrationRepository) Remove(ID string) error {
	return r.db.Delete(&model.Registration{}, ID).Error
}

func (r registrationRepository) UpdatePaymentStatus(ID, status string) error {
	err := r.db.Model(&model.Registration{}).
		Where("id = ?", ID).
		Update("payment_status", status).Error
	return err
}

func (r registrationRepository) GetByEmail(email string) (*model.Registration, error) {
	var registration model.Registration

	result := r.db.Where("email = ?", email).First(&registration)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, errs.ErrInternal.Wrap(result.Error)
		}
	}
	return &registration, nil
}

func (r registrationRepository) GetRegistration(ID string) (*model.Registration, error) {
	var registration model.Registration

	result := r.db.Preload("RegistrationOption").Where("id = ?", ID).First(&registration)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errs.ErrNotFound.Reform("registration not found")
		} else {
			return nil, errs.ErrInternal.Wrap(result.Error)
		}
	}
	return &registration, nil
}

func (r registrationRepository) Create(registration model.Registration) (*model.Registration, error) {
	result := r.db.Create(&registration)
	if result.Error != nil {
		return nil, errs.ErrInternal.Wrap(result.Error)
	}
	return &registration, nil
}

var registrationRepositoryInstance *registrationRepository
var registrationRepositoryOnce sync.Once

func GetRegistrationRepositoryInstance(db *gorm.DB) RegistrationRepository {
	registrationRepositoryOnce.Do(func() {
		registrationRepositoryInstance = &registrationRepository{
			db: db,
		}
	})
	return registrationRepositoryInstance
}
