package models

import (
	"time"
)

type OneIDMerchantSecret struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	MerchantID   string    `gorm:"type:uuid;not null;uniqueIndex:idx_oneid_merchant_secrets_merchant_id" json:"merchant_id"`
	ClientID     string    `gorm:"type:text;not null" json:"client_id"`
	ClientSecret string    `gorm:"type:text;not null" json:"client_secret"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Merchant *Merchant `gorm:"foreignKey:MerchantID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"merchant,omitempty"`
}

func (OneIDMerchantSecret) TableName() string {
	return "oneid_merchant_secret"
}
