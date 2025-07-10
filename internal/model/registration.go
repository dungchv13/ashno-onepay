package model

import (
	"ashno-onepay/internal/errors"
	"database/sql/driver"
	"encoding/json"
)

type Registration struct {
	BaseModel

	RegistrationOptionID string             `json:"registration_option_id"`
	RegistrationOption   RegistrationOption `gorm:"foreignKey:RegistrationOptionID"`

	RegistrationCategory string `gorm:"type:varchar(100)" json:"registration_category" binding:"required"`
	Nationality          string `gorm:"type:varchar(100)" json:"nationality"`
	DoctorateDegree      string `gorm:"type:varchar(100);not null" json:"doctorate_degree" binding:"required"`
	FirstName            string `gorm:"type:varchar(100)" json:"first_name"`
	MiddleName           string `gorm:"type:varchar(100)" json:"middle_name"`
	LastName             string `gorm:"type:varchar(100)" json:"last_name"`
	DateOfBirth          string `json:"date_of_birth"`
	Institution          string `gorm:"type:varchar(255)" json:"institution"`
	Email                string `gorm:"type:varchar(100);not null;uniqueIndex" json:"email" binding:"required,email"`
	PhoneNumber          string `gorm:"type:varchar(20)" json:"phone_number"`
	Sponsor              string `gorm:"type:varchar(255)" json:"sponsor"`

	PaymentStatus    string              `gorm:"type:varchar(50);default:'pending'" json:"payment_status"`
	AccompanyPersons AccompanyPersonList `gorm:"type:jsonb" json:"accompany_persons"`
}

type AccompanyPerson struct {
	FirstName     string `json:"first_name"`
	MiddleName    string `json:"middle_name"`
	LastName      string `json:"last_name"`
	DateOfBirth   string `json:"date_of_birth"`
	PaymentStatus string `json:"payment_status"`
}

type AccompanyPersonDB struct {
	TransactionID  string `gorm:"type:varchar(100)" json:"transaction_id"`
	RegistrationID string `gorm:"type:varchar(100)" json:"registration_id"`
	FirstName      string `gorm:"type:varchar(100)" json:"first_name"`
	MiddleName     string `gorm:"type:varchar(100)" json:"middle_name"`
	LastName       string `gorm:"type:varchar(100)" json:"last_name"`
	DateOfBirth    string `gorm:"type:varchar(100)" json:"date_of_birth"`
}

type AccompanyPersonList []AccompanyPerson

func (a *AccompanyPersonList) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, a)
}

func (a AccompanyPersonList) Value() (driver.Value, error) {
	return json.Marshal(a)
}

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusFail    PaymentStatus = "fail"
	PaymentStatusDone    PaymentStatus = "done"
)

type RegistrationCategory string

const (
	DoctorCategory           RegistrationCategory = "ENT Doctors"
	StudentCategory          RegistrationCategory = "Student & Trainees"
	DoctorAndDinnerCategory  RegistrationCategory = "ENT Doctors + Gala Dinner"
	StudentAndDinnerCategory RegistrationCategory = "Student & Trainees + Gala Dinner"
)

const NationalityVietNam = "vn"

const (
	AccompanyPersonsPaymentStatusPending = "pending"
	AccompanyPersonsPaymentStatusFail    = "fail"
	AccompanyPersonsPaymentStatusDone    = "done"
)
