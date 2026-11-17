package models

import (
	"time"

	"gorm.io/gorm"
)

type EmployeeCardLog struct {
	ID             string             `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	EmployeeCardID string             `gorm:"type:uuid;not null;index" json:"employee_card_id"`
	EmployeeCard   EmployeeCard       `gorm:"foreignKey:EmployeeCardID;constraint:OnDelete:CASCADE" json:"-"`
	Action         EmployeeCardStatus `gorm:"type:varchar(50);not null" json:"action"`
	Remark         string             `gorm:"type:text" json:"remark"`
	PerformedBy    string             `gorm:"type:uuid;index" json:"performed_by"`
	CreatedAt      time.Time          `json:"created_at"`
	DeletedAt      gorm.DeletedAt     `gorm:"index" json:"deleted_at"`
}

func (EmployeeCardLog) TableName() string {
	return "employee_card_logs"
}
