package models

import "time"

type EmployeeCardStatus string

const (
	EmployeeCardStatusPending       EmployeeCardStatus = "pending"
	EmployeeCardStatusRequested     EmployeeCardStatus = "requested"
	EmployeeCardStatusEkycComplete  EmployeeCardStatus = "ekyc_completed"
	EmployeeCardStatusIssued        EmployeeCardStatus = "issued"
	EmployeeCardStatusBlocked       EmployeeCardStatus = "blocked"
	EmployeeCardStatusExpired       EmployeeCardStatus = "expired"
)

type EmployeeCard struct {
	ID              string             `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	MembershipID    string             `gorm:"type:uuid;not null;index" json:"membership_id"`
	Membership      Membership         `gorm:"foreignKey:MembershipID;constraint:OnDelete:CASCADE" json:"-"`
	Status          EmployeeCardStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	OneIDCardURL    string             `gorm:"type:varchar(512)" json:"one_id_card_url"`
	DgepayIDCardURL string             `gorm:"type:varchar(512)" json:"dgepay_id_card_url"`
	DgepayQRCode              string             `gorm:"type:varchar(512)" json:"dgepay_qr_code"`
	DgepayPaymentQRCode       string             `gorm:"type:varchar(512)" json:"dgepay_payment_qr_code"`
	DgepayOneCardThumbnailURL string             `gorm:"type:varchar(512)" json:"dgepay_one_card_thumbnail_url"`
	IsWriteSdk                bool               `gorm:"type:bool;default:false" json:"is_write_sdk"`
	CreatedAt                 time.Time          `json:"created_at"`
	UpdatedAt                 time.Time          `json:"updated_at"`
}
