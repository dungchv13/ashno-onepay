package repository

import (
	errs "ashno-onepay/internal/errors"
	"ashno-onepay/internal/model"
	"errors"
	"sync"

	"gorm.io/gorm"
)

type RegistrationRepository interface {
	Create(registration model.Registration) (*model.Registration, error)
	GetByEmail(email string) (*model.Registration, error)
	GetRegistration(ID string) (*model.Registration, error)
	UpdatePaymentStatus(ID, status string) error
	Remove(ID string) error
	UpdateAccompanyPersonsByID(id string, accompanyPersons model.AccompanyPersonList) error
	SaveAccompanyPersons(persons []model.AccompanyPersonDB) error
	GetAccompanyPersonsByTransactionAndRegistration(transactionID string) ([]model.AccompanyPersonDB, error)
}

type registrationRepository struct {
	db *gorm.DB
}

func (r registrationRepository) Remove(ID string) error {
	return r.db.Where("id = ?", ID).Delete(&model.Registration{}).Error
}

func (r registrationRepository) UpdatePaymentStatus(ID, status string) error {
	err := r.db.Model(&model.Registration{}).
		Where("id = ?", ID).
		Update("payment_status", status).Error
	return err
}

func (r registrationRepository) UpdateAccompanyPersonsByID(id string, accompanyPersons model.AccompanyPersonList) error {
	return r.db.Model(&model.Registration{}).
		Where("id = ?", id).
		Update("accompany_persons", accompanyPersons).Error
}

func (r registrationRepository) SaveAccompanyPersons(persons []model.AccompanyPersonDB) error {
	if len(persons) == 0 {
		return nil
	}
	return r.db.Create(&persons).Error
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

func (r registrationRepository) GetAccompanyPersonsByTransactionAndRegistration(transactionID string) ([]model.AccompanyPersonDB, error) {
	var persons []model.AccompanyPersonDB
	err := r.db.Where("transaction_id = ?", transactionID).Find(&persons).Error
	return persons, err
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
