package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	Id        string    `json:"id" gorm:"column:id;primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime; index; not null;type:timestamp" json:"createdAt"`
	UpdatedAt time.Time `gorm:"autoUpdateTime; not null;type:timestamp" json:"updatedAt"`
}

//nolint:revive
func (baseModel *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if baseModel.Id == "" {
		baseModel.Id = uuid.New().String()
	}
	baseModel.CreatedAt = time.Now().UTC()
	baseModel.UpdatedAt = baseModel.CreatedAt
	return
}

//nolint:revive
func (baseModel *BaseModel) BeforeSave(tx *gorm.DB) (err error) {
	baseModel.UpdatedAt = time.Now().UTC()
	return
}
