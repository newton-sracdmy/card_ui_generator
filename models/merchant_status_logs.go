package models

import (
	"time"

	"gorm.io/gorm"
)

type MerchantStatusLog struct {
	ID          string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	MerchantID  string         `gorm:"type:uuid;not null;index" json:"merchant_id"`
	Merchant    Merchant       `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"merchant"`
	Action      MerchantStatus `gorm:"type:varchar(50);not null" json:"action"`
	Remark      string         `gorm:"type:text" json:"remark"`
	PerformedBy string         `gorm:"type:uuid;index" json:"performed_by"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (MerchantStatusLog) TableName() string {
	return "merchant_status_logs"
}
