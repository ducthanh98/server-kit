package dbgorm

import (
	"gorm.io/gorm"
	"time"
)

type BaseModel struct {
	gorm.Model
	ID        int64          `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type TrackingBaseModel struct {
	gorm.Model
	ID            int64          `json:"id" gorm:"primarykey"`
	CreatedAt     time.Time      `json:"created_at"`
	CreatedUserID int64          `json:"created_user_id"`
	UpdatedAt     time.Time      `json:"updated_at"`
	UpdatedUserID int64          `json:"updated_user_id"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
