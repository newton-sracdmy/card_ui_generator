package dto

type SDKMerchant struct {
	ID           string `json:"id"`
	BusinessName string `json:"business_name"`
}

type MembershipQRQuery struct {
	MerchantID string `query:"merchant_id" validate:"required"`
	RefID      string `query:"ref_id" validate:"required"`
}

type SetIsWriteSdkQuery struct {
	MerchantID string `query:"merchant_id" validate:"required"`
	RefID      string `query:"ref_id" validate:"required"`
	IsWriteSdk bool   `query:"is_write_sdk" validate:"required"`
}
