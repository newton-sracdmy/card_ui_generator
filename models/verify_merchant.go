package models

import (
	"time"

	"gorm.io/gorm"
)

type VerifyMerchant struct {
	ID                         string                 `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	MerchantID                 string                 `gorm:"type:uuid;not null;uniqueIndex" json:"merchant_id"`
	BIN                        bool                   `gorm:"default:false" json:"bin"`
	TINNo                      bool                   `gorm:"default:false" json:"tin_no"`
	BusinessName               bool                   `gorm:"default:false" json:"business_name"`
	PhoneNumber                bool                   `gorm:"default:false" json:"phone_number"`
	Email                      bool                   `gorm:"default:false" json:"email"`
	OneID                      bool                   `gorm:"default:false" json:"one_id"`
	OneIDCategoryID            bool                   `gorm:"default:false" json:"one_id_category_id"`
	Address                    bool                   `gorm:"default:false" json:"address"`
	BusinessCategoryID         bool                   `gorm:"default:false" json:"business_category_id"`
	BusinessServiceTypeID      bool                   `gorm:"default:false" json:"business_service_type_id"`
	BusinessServiceTypeName    bool                   `gorm:"default:false" json:"business_service_type_name"`
	BusinessTypeID             bool                   `gorm:"default:false" json:"business_type_id"`
	ServiceProviderDescription bool                   `gorm:"default:false" json:"service_provider_description"`
	FirstName                  bool                   `gorm:"default:false" json:"first_name"`
	LastName                   bool                   `gorm:"default:false" json:"last_name"`
	Logo                       bool                   `gorm:"default:false" json:"logo"`
	Signature                  bool                   `gorm:"default:false" json:"signature"`
	FaultyDocuments            []MerchantDocument     `gorm:"-" json:"faulty_documents"`
	OneIDServices              []MerchantOneIDService `gorm:"-" json:"one_id_services"`
	Merchant                   Merchant               `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"-"`
	CreatedAt                  time.Time              `json:"created_at"`
	UpdatedAt                  time.Time              `json:"updated_at"`
	DeletedAt                  gorm.DeletedAt         `gorm:"index" json:"deleted_at"`
}

func (VerifyMerchant) TableName() string {
	return "verify_merchants"
}
