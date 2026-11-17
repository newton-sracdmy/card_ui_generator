package models

import "time"


type MerchantOneIDService struct {
	ID                 string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`

	MerchantID         string    `gorm:"type:uuid;not null;index" json:"merchant_id"`
	Merchant           Merchant  `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"-"`

	OneIDServiceID     string    `gorm:"type:varchar(100);not null" json:"one_id_service_id"`
	OneIDService       string `gorm:"type:varchar(255)" json:"one_id_service"`
	IsFault            bool   `gorm:"default:false" json:"is_fault"`

	CreatedAt          time.Time `json:"created_at"`
}