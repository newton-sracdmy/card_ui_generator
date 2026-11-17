package dto

import (
	"employee-card-admin/models"
	"mime/multipart"
)

type RegisterMerchantRequest struct {
	BIN                        string   `form:"bin"`
	TINNo                      string   `form:"tin_no"`
	BusinessName               string   `form:"business_name" validate:"required"`
	PhoneNumber                string   `form:"phone_number" validate:"required"`
	Email                      string   `form:"email" validate:"required,email"`
	OneID                      string   `form:"one_id"`
	OneIDCategoryID            string   `form:"one_id_category_id"`
	OneIDCategoryName          string   `form:"one_id_category_name"`
	BusinessCategory           string   `form:"business_category" validate:"required"`
	BusinessCategoryID         string   `form:"business_category_id"`
	Address                    string   `form:"address" validate:"required"`
	BrandColor                 string   `form:"brand_color" validate:"required"`
	BusinessServiceType        string   `form:"business_service_type"`
	BusinessServiceTypeName    string   `form:"business_service_type_name"`
	ServiceProviderDescription string   `form:"service_provider_description"`
	DialCode                   string   `form:"dial_code"`
	FirstName                  string   `form:"first_name"`
	LastName                   string   `form:"last_name"`
	BusinessType               string   `form:"business_type"`
	OneIDServiceIDs            []string `form:"one_id_service_ids[]"`
	OneIDServices              []string `form:"one_id_service[]"`
	DocumentTypes              []string `form:"document_types"`
	CodeVerifier               string   `form:"code_verifier"`
}

type UpdateMerchantRequest struct {
	BIN                        string                  `form:"bin"`
	TINNo                      string                  `form:"tin_no"`
	BusinessName               string                  `form:"business_name"`
	PhoneNumber                string                  `form:"phone_number"`
	Email                      string                  `form:"email" validate:"omitempty,email"`
	OneID                      string                  `form:"one_id"`
	OneIDCategoryID            string                  `form:"one_id_category_id"`
	OneIDCategoryName          string                  `form:"one_id_category_name"`
	BusinessCategory           string                  `form:"business_category"`
	BusinessCategoryID         string                  `form:"business_category_id"`
	Address                    string                  `form:"address"`
	BrandColor                 string                  `form:"brand_color"`
	BusinessServiceType        string                  `form:"business_service_type"`
	BusinessServiceTypeName    string                  `form:"business_service_type_name"`
	ServiceProviderDescription string                  `form:"service_provider_description"`
	DialCode                   string                  `form:"dial_code"`
	FirstName                  string                  `form:"first_name"`
	LastName                   string                  `form:"last_name"`
	BusinessType               string                  `form:"business_type"`
	DocumentIDs                []string                `form:"document_id[]"`
	DocumentTypes              []string                `form:"document_types"`
	DocumentFiles              []*multipart.FileHeader `form:"document_file[]"`
	MerchantOneIDServiceIDs    []string                `form:"merchant_one_id_service_ids[]"`
	OneIDServiceIDs            []string                `form:"one_id_service_ids[]"`
	OneIDServices              []string                `form:"one_id_service[]"`
}

type MerchantDocumentUpdate struct {
	ID           string
	DocumentType string
	URL          string
}

type MerchantOneIDServiceUpdate struct {
	ID             string
	OneIDServiceID string
	OneIDService   string
}

type ResetMerchantPinRequest struct {
	CurentPIN string `json:"current_pin" validate:"required"`
	PIN       string `json:"pin" validate:"required"`
}

type VerifyMerchantRegistrationRequest struct {
	BIN                        bool     `json:"bin"`
	TINNo                      bool     `json:"tin_no"`
	BusinessName               bool     `json:"business_name"`
	PhoneNumber                bool     `json:"phone_number"`
	Email                      bool     `json:"email"`
	OneID                      bool     `json:"one_id"`
	OneIDCategoryID            bool     `json:"one_id_category_id"`
	Address                    bool     `json:"address"`
	BusinessCategoryID         bool     `json:"business_category_id"`
	BusinessServiceTypeID      bool     `json:"business_service_type_id"`
	BusinessServiceTypeName    bool     `json:"business_service_type_name"`
	BusinessTypeID             bool     `json:"business_type_id"`
	ServiceProviderDescription bool     `json:"service_provider_description"`
	FirstName                  bool     `json:"first_name"`
	LastName                   bool     `json:"last_name"`
	Logo                       bool     `json:"logo"`
	Signature                  bool     `json:"signature"`
	MerchantOneIDServiceIDs    []string `json:"merchant_one_id_service_ids"`
	Documents                  []string `json:"documents"`
}

type ChangeMerchantStatusRequest struct {
	Status models.MerchantStatus `json:"status" validate:"required"`
	Remark string                `json:"remark" validate:"required"`

	VerifyMerchantRegistrationRequest
}

type ChangeMerchantStatusResponse struct {
	ID     string                `json:"id"`
	Status models.MerchantStatus `json:"status"`
	Remark string                `json:"remark"`
}

type RegisterAdminResponse struct {
	Message string `json:"message"`
}