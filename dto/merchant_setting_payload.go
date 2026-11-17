package dto

type CreateMerchantSettingRequest struct {
	BrandColor          string  `json:"brand_color" form:"brand_color"`
	Logo                string  `json:"logo" form:"logo"`
	Signature           string  `json:"signature" form:"signature"`
	Address             string  `json:"address" form:"address"`
	OptionalPhoneNumber *string `json:"optional_phone_number,omitempty" form:"optional_phone_number"`
	Fax                 *string `json:"fax,omitempty" form:"fax"`
	WebsiteURL          *string `json:"website_url,omitempty" form:"website_url"`
}

type UpdateMerchantSettingRequest struct {
	BusinessName        *string `json:"business_name,omitempty" form:"business_name"`
	BrandColor          *string `json:"brand_color,omitempty" form:"brand_color"`
	Logo                *string `json:"logo,omitempty" form:"logo"`
	Signature           *string `json:"signature,omitempty" form:"signature"`
	Address             *string `json:"address,omitempty" form:"address"`
	OptionalPhoneNumber *string `json:"optional_phone_number,omitempty" form:"optional_phone_number"`
	Fax                 *string `json:"fax,omitempty" form:"fax"`
	WebsiteURL          *string `json:"website_url,omitempty" form:"website_url"`
	CardExpiryTime      *int64  `json:"card_expiry_time,omitempty" form:"card_expiry_time"`
}

type MerchantSettingResponse struct {
	ID                  string  `json:"id"`
	BusinessName        string  `json:"business_name,omitempty"`
	BrandColor          string  `json:"brand_color"`
	Logo                string  `json:"logo"`
	Signature           string  `json:"signature"`
	Address             string  `json:"address"`
	OptionalPhoneNumber *string `json:"optional_phone_number,omitempty"`
	Fax                 *string `json:"fax,omitempty"`
	WebsiteURL          *string `json:"website_url,omitempty"`
	CardExpiryTime      *int64  `json:"card_expiry_time,omitempty"`
	CreatedAt           int64   `json:"created_at"`
	UpdatedAt           int64   `json:"updated_at"`
}
