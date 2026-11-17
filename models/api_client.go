package models

import (
	"time"
)

// APIClientStatus enum
type APIClientStatus string

const (
	APIClientActive   APIClientStatus = "active"
	APIClientInactive APIClientStatus = "inactive"
)

type APIClient struct {
	ID               string          `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	ClientID         string          `gorm:"type:varchar(255);not null;uniqueIndex" json:"client_id"`
	ClientSecretHash string          `gorm:"type:text;not null" json:"client_secret_hash"` // hashed secret
	Status           APIClientStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (APIClient) TableName() string {
	return "api_clients"
}