package models

import "time"

type ThirdPartyAPIEntityType string

const (
	ThirdPartyAPIEntityTypeEmployee ThirdPartyAPIEntityType = "employee"
	ThirdPartyAPIEntityTypeMerchant ThirdPartyAPIEntityType = "merchant"
)

type ThirdPartyAPILog struct {
	ID   string                  `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	Type ThirdPartyAPIEntityType `gorm:"type:varchar(20);not null;uniqueIndex:idx_third_party_api_logs_type_reference_id" json:"type"`

	ReferenceID string `gorm:"column:reference_id;type:uuid;not null;uniqueIndex:idx_third_party_api_logs_type_reference_id" json:"reference_id"`

	CreateUserEmpCard     bool    `gorm:"default:false" json:"create_user_emp_card"`
	CreateUserEmpCardID   string  `gorm:"type:text" json:"-"`

	GenerateVDSDocument     bool    `gorm:"default:false" json:"generate_vds_document"`
	GenerateVDSDocumentData JSONMap `gorm:"type:jsonb" json:"-"`

	SyncDocument bool `gorm:"default:false" json:"sync_document"`

	DgepayMerchantRegister   bool   `gorm:"default:false" json:"dgepay_Merchant_register"`
	DgepayMerchantID         string `gorm:"type:text" json:"-"`
	CreateBusinessMaster bool   `gorm:"default:false" json:"create_business_master"`
	DgepayBusinessID     string `gorm:"type:text" json:"-"`

	CreateServiceProvider bool   `gorm:"default:false" json:"create_service_provider"`
	ServiceProviderID     string `gorm:"type:text" json:"-"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ThirdPartyAPILog) TableName() string {
	return "third_party_api_logs"
}
