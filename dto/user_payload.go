// dto/user_payload.go
package dto

type UserPayload struct {
	MerchantID string `json:"merchant_id" validate:"required"`
	Phone      string `json:"phone" validate:"required"`
}

type LoginPayload struct {
	Phone string `json:"phone" validate:"required"`
	PIN   string `json:"pin" validate:"required"`
}

type InviteUserPayload struct {
	EmployeeID string `json:"employee_id" validate:"required"` 
}

type DGEPayUserResponse struct {
	UserID string `json:"user_id"`
}
