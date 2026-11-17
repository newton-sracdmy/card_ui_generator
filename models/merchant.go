package models

import (
	"database/sql/driver"
	"employee-card-admin/common/enums"
	"time"

	"gorm.io/gorm"
)

// MerchantStatus enum
type MerchantStatus string

const (
	Pending     MerchantStatus = "pending"
	Active      MerchantStatus = "active"
	Rejected    MerchantStatus = "rejected"
	Resubmitted MerchantStatus = "resubmitted"
)

func (ms MerchantStatus) Value() (driver.Value, error) {
	return string(ms), nil
}

func (ms *MerchantStatus) Scan(value interface{}) error {
	if value == nil {
		*ms = ""
		return nil
	}
	*ms = MerchantStatus(value.(string))
	return nil
}

type Merchant struct {
	ID                         string                 `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	BIN                        string                 `json:"bin"`
	TINNo                      string                 `gorm:"type:varchar(255)" json:"tin_no"`
	BusinessName               string                 `gorm:"uniqueIndex;not null" json:"business_name"`
	Email                      string                 `gorm:"uniqueIndex;not null" json:"email"`
	OneID                      string                 `gorm:"type:varchar(255)" json:"one_id"`
	OneIDCategoryID            string                 `gorm:"type:uuid" json:"one_id_category_id"`
	OneIDCategoryName          string                 `gorm:"type:varchar(255)" json:"one_id_category_name"`
	BusinessCategory           string                 `gorm:"type:varchar(255);not null" json:"business_category"`
	BusinessCategoryID         string                 `gorm:"type:uuid" json:"business_category_id"`
	BusinessServiceType        string                 `gorm:"type:uuid" json:"business_service_type"`
	BusinessServiceTypeName    string                 `gorm:"type:varchar(255)" json:"business_service_type_name"`
	ServiceProviderDescription string                 `gorm:"type:text" json:"service_provider_description"`
	DialCode                   string                 `gorm:"type:varchar(10)" json:"dial_code"`
	FirstName                  string                 `gorm:"type:varchar(150)" json:"first_name"`
	LastName                   string                 `gorm:"type:varchar(150)" json:"last_name"`
	BusinessType               string                 `gorm:"type:varchar(150)" json:"business_type"`
	Phone                      string                 `gorm:"uniqueIndex;not null" json:"phone"`
	PIN                        string                 `gorm:"not null" json:"-"`
	IsChangePIN                bool                   `gorm:"column:is_change_pin;not null;default:false" json:"is_change_pin"`
	UserType                   enums.UserType         `gorm:"type:varchar(20);default:'merchant'" json:"user_type"`
	ServiceProviderID          string                 `gorm:"column:service_provider_id;type:text;index:idx_merchants_service_provider_id" json:"service_provider_id"`
	DGEPayMerchantID           string                 `gorm:"column:dge_pay_merchant_id;type:text" json:"dge_pay_merchant_id"`
	DGEPayBusinessID           string                 `gorm:"column:dge_pay_business_id;type:text" json:"dge_pay_business_id"`
	OneIDRefreshToken          string                 `gorm:"column:one_id_refresh_token;type:text" json:"-"`
	Status                     MerchantStatus         `gorm:"type:varchar(20);default:'pending'" json:"status"`
	Documents                  []MerchantDocument     `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"documents"`
	Settings                   *MerchantSetting       `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"settings,omitempty"`
	OneIDServices              []MerchantOneIDService `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"one_id_services,omitempty"`
	CreatedAt                  time.Time              `json:"created_at"`
	UpdatedAt                  time.Time              `json:"updated_at"`
	DeletedAt                  gorm.DeletedAt         `gorm:"index" json:"deleted_at"`
}

func (Merchant) TableName() string {
	return "merchants"
}

type MerchantDocument struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	MerchantID   string    `gorm:"type:uuid;not null;index" json:"merchant_id"`
	DocumentType string    `gorm:"not null" json:"document_type"`
	URL          string    `gorm:"not null" json:"url"`
	Merchant     Merchant  `gorm:"foreignKey:MerchantID;constraint:OnDelete:CASCADE" json:"-"`
	IsFault      bool      `gorm:"default:false" json:"is_fault"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (MerchantDocument) TableName() string {
	return "merchant_documents"
}
