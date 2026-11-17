package dto

import "encoding/json"

type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	TokenType        string `json:"token_type"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
}

type ExchangeCodeRequest struct {
	Code         string `json:"code"`
	CodeVerifier string `json:"code_verifier"`
}

type OneIDResponse struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

type CreateServiceProviderRequest struct {
	Name         string   `json:"name"`
	Email        string   `json:"email"`
	TINNo        string   `json:"tin_no"`
	Phone        string   `json:"phone"`
	CategoryID   string   `json:"category_id"`
	Description  string   `json:"description"`
	CreateClient bool     `json:"create-client"`
	ServiceIDs   []string `json:"service_ids"`
}

type ServiceData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type CreateServiceProviderData struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Email        string        `json:"email"`
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	TINNo        string        `json:"tin_no"`
	Phone        string        `json:"phone"`
	CategoryID   string        `json:"category_id"`
	CategoryName string        `json:"category_name"`
	Description  string        `json:"description"`
	Services     []ServiceData `json:"services"`
	IsActive     bool          `json:"is_active"`
	Status       string        `json:"status"`
	IsOnboarded  bool          `json:"is_onboarded"`
	CreatedAt    string        `json:"created_at"`
	UpdatedAt    string        `json:"updated_at"`
}

type CategoryData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type CategoriesResponseData struct {
	TotalCategories int            `json:"total_categories"`
	CurrentPage     int            `json:"current_page"`
	TotalPages      int            `json:"total_pages"`
	Categories      []CategoryData `json:"categories"`
}

type DocumentCategoryData struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Logo      *string `json:"logo"`
	IsActive  bool    `json:"is_active"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type ServiceProviderCategoryData struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GetServicesResponseService struct {
	ID                      string                      `json:"id"`
	Name                    string                      `json:"name"`
	Logo                    *string                     `json:"logo"`
	DocumentCategory        DocumentCategoryData        `json:"document_category"`
	ServiceProviderCategory ServiceProviderCategoryData `json:"service_provider_category"`
	IsActive                bool                        `json:"is_active"`
	CreatedAt               string                      `json:"created_at"`
	UpdatedAt               string                      `json:"updated_at"`
}

type GetServicesResponseData struct {
	TotalServices int                          `json:"total_services"`
	CurrentPage   int                          `json:"current_page"`
	TotalPages    int                          `json:"total_pages"`
	Services      []GetServicesResponseService `json:"services"`
}

type ServiceListItem struct {
	ServiceID   string `json:"service_id"`
	ServiceName string `json:"service_name"`
}

type InviteUserRequest struct {
	MerchantID  string `json:"merchant_id,omitempty"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	ProviderID  string `json:"provider_id,omitempty"`
	RefID       string `json:"ref_id,omitempty"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

type InviteUserData struct {
	Email              string `json:"email"`
	ProviderID         string `json:"provider_id"`
	HasAccount         bool   `json:"has_account"`
	AssociationCreated bool   `json:"association_created"`
	Message            string `json:"message"`
	RefID              string `json:"ref_id"`
	RedirectURI        string `json:"redirect_uri"`
	OneID              string `json:"one_id"`
}

type RegisterServiceProviderUserRequest struct {
	MerchantID  string `json:"-"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
	ProviderID  string `json:"provider_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
}

type RegisterServiceProviderUserData struct {
	UserID       string `json:"user_id"`
	OneID        string `json:"one_id"`
	Email        string `json:"email"`
	PhoneNumber  string `json:"phone_number"`
	ProviderID   string `json:"provider_id"`
	RefID        string `json:"ref_id"`
	EkycVerified bool   `json:"ekyc_verified"`
	// AlreadyExists is true when OneID returned an existing account instead of creating one.
	// RefID is empty when that account is not linked to the requesting provider.
	AlreadyExists bool `json:"already_exists"`
}

type CreateEKYCSessionRequest struct {
	MerchantID             string `json:"merchant_id,omitempty"`
	Phone                  string `json:"phone" validate:"required"`
	Email                  string `json:"email" validate:"required,email"`
	IdentityType           string `json:"identity_type" validate:"required"`
	RedirectURI            string `json:"redirect_uri" validate:"required,url"`
	SessionValiditySeconds int    `json:"session_validity_seconds" validate:"required,gt=0"`
}

type CreateEKYCSessionData struct {
	SessionID         string `json:"session_id"`
	EKYCURL           string `json:"ekyc_url"`
	ExpiresAt         string `json:"expires_at"`
	Status            string `json:"status"`
	Verified          bool   `json:"verified"`
	IsExistingSession bool   `json:"is_existing_session"`
}

type InviteUserResponse struct {
	InviteData *InviteUserData        `json:"invite_data,omitempty"`
	EKYCData   *CreateEKYCSessionData `json:"ekyc_data,omitempty"`
}

type QRMetadata map[string]map[string]string

type GenerateVDSRequest struct {
	MerchantID string     `json:"merchant_id,omitempty"`
	OneID      string     `json:"one_id,omitempty"`
	RefID      string     `json:"ref_id,omitempty"`
	ProviderID string     `json:"provider_id,omitempty"`
	ServiceID  string     `json:"service_id,omitempty"`
	QRMetadata QRMetadata `json:"qrMetadata,omitempty"`
}

type GenerateVDSData struct {
	DocumentID       string `json:"document_id"`
	ExtractedQRImage string `json:"extractedQrImage"`
}

type VerifyVDSRequest struct {
	QRRaw string `json:"qr_raw"`
}

type DecodedPayloadField struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type CertificateInfo struct {
	Subject       string `json:"subject"`
	Issuer        string `json:"issuer"`
	SerialDecimal string `json:"serial_decimal"`
	SerialHex     string `json:"serial_hex"`
	PublicKeyType string `json:"public_key_type"`
	NotBefore     string `json:"not_before"`
	NotAfter      string `json:"not_after"`
}

type DocumentInfo struct {
	DocumentExists bool   `json:"document_exists"`
	DocumentStatus string `json:"document_status"`
	IsDeleted      bool   `json:"is_deleted"`
	CategoryName   string `json:"category_name"`
	DocumentType   string `json:"document_type"`
	IssuedAt       string `json:"issued_at"`
}

type VerifyVDSData struct {
	VerificationStatus bool                           `json:"verification_status"`
	SignatureValid     bool                           `json:"signature_valid"`
	SerialMatch        bool                           `json:"serial_match"`
	DecodedPayload     map[string]DecodedPayloadField `json:"decoded_payload"`
	DocumentInfo       DocumentInfo                   `json:"document_info"`
	CertificateInfo    CertificateInfo                `json:"certificate_info"`
}

type SyncDocumentData struct {
	ID                string  `json:"id"`
	UserID            string  `json:"user_id"`
	CategoryID        string  `json:"category_id"`
	CategoryName      *string `json:"category_name"`
	RequestID         *string `json:"request_id"`
	IdentityDocType   *string `json:"identity_document_type"`
	Extension         string  `json:"extension"`
	VdsQR             *string `json:"vds_qr"`
	Content           string  `json:"content"`
	VdsPassportHashQR *string `json:"vds_passport_hash_qr"`
	StampedDocument   *string `json:"stamped_document"`
	ServiceName       *string `json:"service_name"`
	ProviderName      *string `json:"provider_name"`
	Metadata          *string `json:"metadata"`
	CreatedAt         string  `json:"created_at"`
	UpdatedAt         string  `json:"updated_at"`
	Status            string  `json:"status"`
}

type SyncDocumentRequest struct {
	DocumentType string `json:"document_type" validate:"required"`
	DocumentID   string `json:"document_id" validate:"required"`
	RefID        string `json:"ref_id,omitempty"`
}
type OneIDUserInfoResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    OneIDUserInfoData `json:"data"`
}

type OneIDUserInfoData struct {
	OneID           string      `json:"one_id"`
	Email           string      `json:"email"`
	FirstName       string      `json:"first_name"`
	LastName        string      `json:"last_name"`
	PhoneNumber     string      `json:"phone_number"`
	EmailVerified   bool        `json:"email_verified"`
	PhoneVerified   bool        `json:"phone_verified"`
	EkycVerified    bool        `json:"ekyc_verified"`
	Country         string      `json:"country"`
	RefID           interface{} `json:"ref_id"`
	ProviderID      string      `json:"provider_id"`
	ProviderName    string      `json:"provider_name"`
	ProfilePhotoB64 string      `json:"profile_photo_b64"`
}

type OneIDOpenIDUserInfo map[string]interface{}

type FaceMatchResponse struct {
	Similarity float64 `json:"similarity"`
	Threshold  float64 `json:"threshold"`
	Verified   bool    `json:"verified"`
}

type UpdateUserCardRequest struct {
	Email      string `json:"email"`
	FrontImage string `json:"front_image"`
	BackImage  string `json:"back_image"`
	Thumbnail  string `json:"thumbnail"`
}

