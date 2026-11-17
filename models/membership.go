package models

import "time"

type MembershipStatus string

const (
	MembershipActive    MembershipStatus = "active"
	MembershipSuspended MembershipStatus = "suspended"
	MembershipRemoved   MembershipStatus = "removed"
)

// Onboard statuses for memberships that are onboarded automatically (Yogi).
// An empty status means onboarding has not finished yet.
const (
	MembershipOnboardCompleted   = "completed"
	MembershipOnboardNeedsReview = "needs_review"
)

type Membership struct {
	ID              string           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	EmployeeID      string           `gorm:"type:uuid;not null;uniqueIndex:idx_emp_merchant" json:"employee_id"`
	Employee        Employee         `gorm:"foreignKey:EmployeeID" json:"-"`
	MerchantID      string           `gorm:"type:uuid;not null;uniqueIndex:idx_emp_merchant" json:"merchant_id"`
	Merchant        Merchant         `gorm:"foreignKey:MerchantID" json:"-"`
	Designation     string           `gorm:"not null" json:"designation"`
	RefID           string           `gorm:"type:varchar(100);index" json:"ref_id"`
	Status          MembershipStatus `gorm:"type:varchar(20);default:'active'" json:"status"`
	OneIDServiceID  string           `gorm:"type:varchar(100)" json:"one_id_service_id"`
	MetaData        JSONMap          `gorm:"type:jsonb" json:"meta_data"`
	BulkUploadJobID *string          `gorm:"type:uuid;index" json:"-"`
	OnboardStatus   string           `gorm:"type:varchar(30);not null;default:''" json:"onboard_status,omitempty"`
	OnboardError    string           `gorm:"type:text;not null;default:''" json:"-"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}
