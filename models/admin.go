package models

import (
	"database/sql/driver"
	"employee-card-admin/common/enums"
	"time"
)

type AdminStatus string

const (
	AdminActive   AdminStatus = "active"
	AdminInactive AdminStatus = "inactive"
)

func (as AdminStatus) Value() (driver.Value, error) {
	return string(as), nil
}

func (as *AdminStatus) Scan(value interface{}) error {
	if value == nil {
		*as = ""
		return nil
	}

	*as = AdminStatus(value.(string))
	return nil
}

type Admin struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Name      string         `gorm:"type:varchar(255);not null" json:"name"`
	Phone     string         `gorm:"type:varchar(30);not null;uniqueIndex" json:"phone"`
	Email     string         `gorm:"type:varchar(255);not null;uniqueIndex" json:"email"`
	Password  string         `gorm:"type:text;not null" json:"-"`
	UserType  enums.UserType `gorm:"type:varchar(20);not null;default:'admin'" json:"user_type"`
	Status    AdminStatus    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (Admin) TableName() string {
	return "admins"
}
