package models

import (
	"time"

	"gorm.io/gorm"
)

type MerchantSetting struct {
	ID                  string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	MerchantID          string         `gorm:"type:uuid;not null;index" json:"merchant_id"`
	BrandColor          string         `gorm:"type:varchar(255);not null" json:"brand_color"`
	Logo                string         `gorm:"type:varchar(255);not null" json:"logo"`
	Signature           string         `gorm:"type:varchar(255);not null" json:"signature"`
	Address             string         `gorm:"not null" json:"address"`
	OptionalPhoneNumber *string        `gorm:"column:optional_phone_number;type:varchar(255)" json:"optional_phone_number,omitempty"`
	Fax                 *string        `gorm:"type:varchar(255)" json:"fax,omitempty"`
	WebsiteURL          *string        `gorm:"type:varchar(255)" json:"website_url,omitempty"`
	CardExpiryTime      time.Time      `gorm:"not null;default:now() + interval '3 years'" json:"card_expiry_time"`
	Merchant            Merchant       `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (MerchantSetting) TableName() string {
	return "merchant_settings"
}
