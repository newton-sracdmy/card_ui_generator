package dto

import (
	"employee-card-admin/models"
)

type ClientAPIEmployeeRegisterRequest struct {
	ServiceProviderID string         `json:"service_provider_id" validate:"required"`
	ServiceID         string         `json:"service_id" validate:"required"`
	UserName          string         `json:"user_name" validate:"required"`
	Designation       string         `json:"designation" validate:"required"`
	Department        string         `json:"department"`
	Email             string         `json:"email" validate:"required,email"`
	Phone             string         `json:"phone" validate:"required"`
	RefID             string         `json:"ref_id" validate:"required"`
	AvatarURL         string         `json:"avatar_url"`
	AvaterURL         string         `json:"avater_url,omitempty"`
	NID               string         `json:"nid"`
	DateOfBirth       string         `json:"date_of_birth"`
	Program           string         `json:"program"`
	BloodGroup        string         `json:"blood_group"`
	OneID             string         `json:"one_id"`
	OneIDRefID        string         `json:"one_id_ref_id"`
	MetaData          models.JSONMap `json:"meta_data,omitempty"`
}

type ClientAPIEmployeeRegisterResponse struct {
	MembershipID string `json:"membership_id"`
	JobID        string `json:"job_id,omitempty"`
}

type ClientAPIGetEmployeeCardsRequest struct {
	Email string `json:"email" validate:"required"`
}

type ClientAPIGetEmployeeCardsResponse struct {
	BackCardURL               string `json:"back_card_url"`
	FrontCardURL              string `json:"front_card_URL"`
	DgepayOneCardThumbnailURL string `json:"dgepayone_card_thumnail_url"`
}
