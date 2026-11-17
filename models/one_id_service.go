package models

import (
	"time"

	"gorm.io/gorm"
)

type OneIDServiceItem struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ServiceID   string         `gorm:"type:uuid;not null;uniqueIndex" json:"service_id"`
	ServiceName string         `gorm:"type:varchar(255);not null" json:"service_name"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (OneIDServiceItem) TableName() string {
	return "one_id_services"
}
