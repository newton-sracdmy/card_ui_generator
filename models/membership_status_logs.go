package models

import (
	"time"

	"gorm.io/gorm"
)

type MembershipStatusLog struct {
	ID           string           `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	MembershipID string           `gorm:"type:uuid;not null;index" json:"membership_id"`
	Membership   Membership       `gorm:"foreignKey:MembershipID;constraint:OnDelete:CASCADE" json:"-"`
	Action       MembershipStatus `gorm:"type:varchar(50);not null" json:"action"`
	Remark       string           `gorm:"type:text" json:"remark"`
	PerformedBy  string           `gorm:"type:uuid;index" json:"performed_by"`
	CreatedAt    time.Time        `json:"created_at"`
	DeletedAt    gorm.DeletedAt   `gorm:"index" json:"deleted_at"`
}

func (MembershipStatusLog) TableName() string {
	return "membership_status_logs"
}
