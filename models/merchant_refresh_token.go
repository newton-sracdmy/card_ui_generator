package models

import (
	"time"
)

type MerchantRefreshToken struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	RefreshToken string    `gorm:"type:text;not null" json:"refresh_token"`
	CodeVerifier string    `gorm:"type:text;not null" json:"code_verifier"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (MerchantRefreshToken) TableName() string {
	return "merchant_refresh_tokens"
}
