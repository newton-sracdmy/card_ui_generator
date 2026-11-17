package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type EmployeeStatus string

const (
	EmployeeActive    EmployeeStatus = "active"
	EmployeeDeleted   EmployeeStatus = "deleted"
)

func (es EmployeeStatus) Value() (driver.Value, error) {
	return string(es), nil
}

func (es *EmployeeStatus) Scan(value interface{}) error {
	if value == nil {
		*es = ""
		return nil
	}
	*es = EmployeeStatus(value.(string))
	return nil
}

type JSONMap map[string]interface{}

func (jm JSONMap) Value() (driver.Value, error) {
	if len(jm) == 0 {
		return nil, nil
	}
	return json.Marshal(jm)
}

func (jm *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*jm = make(JSONMap)
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		if s, ok := value.(string); ok {
			b = []byte(s)
		} else {
			return errors.New("type assertion to []byte failed")
		}
	}
	return json.Unmarshal(b, jm)
}

type Employee struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	UserName       string         `gorm:"not null" json:"user_name"`
	Email          string         `gorm:"not null;unique;index:idx_employees_email" json:"email"`
	Phone          string         `gorm:"not null;unique" json:"phone"`
	Status         EmployeeStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	OneID          string        `gorm:"type:varchar(100)" json:"one_id"`
	OneIDRefID     string        `gorm:"type:varchar(255)" json:"one_id_ref_id"`
	AvatarURL      string         `gorm:"type:varchar(512)" json:"avatar_url"`
	DgepayUserID   string         `gorm:"type:varchar(255)" json:"dgepay_user_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

func (Employee) TableName() string {
	return "employees"
}
