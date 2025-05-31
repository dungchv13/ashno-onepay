package model

type User struct {
	BaseModel
	Email       string `json:"email" gorm:"email; not null; uniqueIndex"`
	Password    string `json:"password" gorm:"password"`
	Phone       string `json:"phone" gorm:"phone"`
	Nationality string `json:"nationality" gorm:"nationality"`
	FirstName   string `json:"first_name" gorm:"first_name"`
	MiddleName  string `json:"middle_name" gorm:"middle_name"`
	LastName    string `json:"last_name" gorm:"last_name"`
	DateOfBirth string `json:"date_of_birth" gorm:"date_of_birth"`
	Institution string `json:"institution" gorm:"institution"`
	SponsorBy   string `json:"sponsor_by" gorm:"sponsor_by"`
}

func (*User) TableName() string {
	return "users"
}
