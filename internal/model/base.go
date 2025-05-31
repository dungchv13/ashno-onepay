package model

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	Id        uint      `json:"id" gorm:"primary_key;auto_increment;column:id"`
	CreatedAt time.Time `gorm:"autoCreateTime; index; not null;type:timestamp" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime; not null;type:timestamp" json:"updatedAt"`
}

//nolint:revive
func (baseModel *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	baseModel.CreatedAt = time.Now().UTC()
	baseModel.UpdatedAt = baseModel.CreatedAt
	return
}

//nolint:revive
func (baseModel *BaseModel) BeforeSave(tx *gorm.DB) (err error) {
	baseModel.UpdatedAt = time.Now().UTC()
	return
}
