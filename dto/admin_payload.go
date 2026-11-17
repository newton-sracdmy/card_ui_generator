package dto

import (
	"employee-card-admin/common/enums"
	"employee-card-admin/models"
)

type AdminSignupRequest struct {
	Name     string `json:"name" validate:"required"`
	Phone    string `json:"phone" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type AdminSignupResponse struct {
	ID       string             `json:"id"`
	Name     string             `json:"name"`
	Phone    string             `json:"phone"`
	Email    string             `json:"email"`
	UserType enums.UserType     `json:"user_type"`
	Status   models.AdminStatus `json:"status"`
}

type AdminSigninRequest struct {
	Phone    string `json:"phone" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AdminSigninResponse struct {
	AccessToken string               `json:"access_token"`
	Admin       AdminSignupResponse `json:"admin"`
}
