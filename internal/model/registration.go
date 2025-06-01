package model

type Registration struct {
	BaseModel

	RegistrationOptionID string             `json:"-"`
	RegistrationOption   RegistrationOption `gorm:"foreignKey:RegistrationOptionID" json:"-"`

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

	PaymentStatus string `gorm:"type:varchar(50);default:'pending'" json:"payment_status"`
}

type PaymentStatus string
type RegistrationCategory string

const (
	PaymentStatusPending PaymentStatus        = "pending"
	PaymentStatusDone    PaymentStatus        = "done"
	NationalityVietNam                        = "vn"
	DoctorCategory       RegistrationCategory = "ENT Doctors"
	StudentCategory      RegistrationCategory = "Student & Trainees"
	DinnerCategory       RegistrationCategory = "Chairman & Speaker"
)
