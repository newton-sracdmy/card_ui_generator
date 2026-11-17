package dto

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginData struct {
	Token      string `json:"token"`
	ExpireTime int64  `json:"expire_time"`
}

type LoginResponse struct {
	Success int       `json:"success"`
	Error   []string  `json:"error"`
	Data    LoginData `json:"data"`
}

type OneIDDetails struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	OneID             string `json:"one_id"`
	AccessToken       string `json:"access_token"`
	PreferredUsername string `json:"preferred_username"`
	FamilyName        string `json:"family_name"`
}

type CustomerRegisterRequest struct {
	DialCode     string       `json:"dial_code"`
	PhoneNumber  string       `json:"phone_number"`
	FirstName    string       `json:"first_name"`
	LastName     string       `json:"last_name"`
	BirthDate    int64        `json:"birth_date"`
	NID          string       `json:"nid"`
	Email        string       `json:"email"`
	OneIDDetails OneIDDetails `json:"one_id_details"`
}

type TokenData struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type CustomerRegisterData struct {
	Message     string    `json:"message"`
	EventLogsID string    `json:"event_logs_id"`
	Token       TokenData `json:"token"`
	UserID      string    `json:"user_id"`
}

type CustomerRegisterResponse struct {
	Success int                  `json:"success"`
	Error   []string             `json:"error"`
	Data    CustomerRegisterData `json:"data,omitempty"`
}

type BusinessMasterRequest struct {
	BusinessName          string `json:"business_name"`
	BusinessType          string `json:"business_type"`
	BusinessCategoryID    string `json:"business_category_id"`
	MonthlyTurnover       int    `json:"monthly_turnover"`
	VSChallanRegistration bool   `json:"vschallan_registration"`
	UserID                string `json:"user_id"`
}

type BusinessMasterData struct {
	Message          string `json:"message"`
	EventLogsID      string `json:"event_logs_id"`
	BusinessMasterID string `json:"business_master_id"`
}

type BusinessMasterResponse struct {
	Success int                `json:"success"`
	Error   []string           `json:"error"`
	Data    BusinessMasterData `json:"data,omitempty"`
}

type UserRegisterRequest struct {
	DialCode    string  `json:"dial_code"`
	PhoneNumber string  `json:"phone_number"`
	UserType    int     `json:"user_type"`
	FirstName   string  `json:"first_name"`
	LastName    *string `json:"last_name,omitempty"`
}

type UserEmpCardRequest struct {
	CardNumber string `json:"card_number"`
	UserID     string `json:"user_id"`
	CardType   string `json:"card_type"`
	MerchantID string `json:"merchant_id"`
}

type GenericResponse struct {
	Success int      `json:"success"`
	Error   []string `json:"error"`
	Data    any      `json:"data,omitempty"`
}

type CreateUserEmpCardResponse struct {
	Success int      `json:"success"`
	Error   []string `json:"error"`
	Data    struct {
		ID string `json:"id"`
	} `json:"data,omitempty"`
}

type EmpCardBrandingMetadata struct {
	FrontImageURL     string `json:"front_image_url"`
	ThumbnailImageURL string `json:"thumbnail_image_url"`
	VDSQRCodeURL      string `json:"vds_qr_code_url"`
}

type UpdateUserEmpCardBrandingMetadataRequest struct {
	ID              string                  `json:"id"`
	BrandingMetadata EmpCardBrandingMetadata `json:"branding_metadata"`
}
