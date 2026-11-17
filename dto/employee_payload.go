package dto

import "employee-card-admin/models"

type CreateEmployeeRequest struct {
	UserName       string         `json:"user_name" validate:"required"`
	Designation    string         `json:"designation" validate:"required"`
	Email          string         `json:"email" validate:"required,email"`
	Phone          string         `json:"phone" validate:"required"`
	Status         string         `json:"status" validate:"required"`
	RefID          string         `json:"ref_id" validate:"required"`
	AvatarURL      string         `json:"avatar_url" validate:"required"`
	OneIDServiceID string         `json:"one_id_service_id"`
	MetaData       models.JSONMap `json:"meta_data"`
}

type DeleteEmployeeResponse struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

type GenerateCardRequest struct {
	MembershipID string `json:"membership_id" validate:"required"`
}

type BulkGenerateCardsRequest struct {
	MembershipIDs []string `json:"membership_ids" validate:"required,min=1"`
}

type BulkGenerateCardsResponse struct {
	Message      string                   `json:"message"`
	Results      []BulkGenerateCardResult `json:"results,omitempty"`
	SuccessCount int                      `json:"success_count"`
	FailedCount  int                      `json:"failed_count"`
}

type BulkGenerateCardResult struct {
	MembershipID string                `json:"membership_id"`
	Result       *GenerateCardResponse `json:"result,omitempty"`
	Error        string                `json:"error,omitempty"`
}

type GeneratedCardImage struct {
	Name      string `json:"name"`
	MimeType  string `json:"mime_type"`
	URL       string `json:"url"`
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
}

type GenerateCardResponse struct {
	EmployeeID    string              `json:"employee_id"`
	OneIDCard     GeneratedCardImage  `json:"one_id_card"`
	PaymentCard   GeneratedCardImage  `json:"payment_card"`
	ThumbnailCard *GeneratedCardImage `json:"thumbnail_card,omitempty"`
	Success       bool                `json:"success"`
	Message       string              `json:"message"`
}

type EmployeeCardResponse struct {
	UserName                  string `json:"user_name"`
	Designation               string `json:"designation"`
	Phone                     string `json:"phone"`
	Email                     string `json:"email"`
	AvatarURL                 string `json:"avatar_url"`
	OneIDCardURL              string `json:"one_id_card_url"`
	DgepayIDCardURL           string `json:"dgepay_id_card_url"`
	DgepayQRCode              string `json:"dgepay_qr_code"`
	DgepayPaymentQRCode       string `json:"dgepay_payment_qr_code"`
	DgepayOneCardThumbnailURL string `json:"dgepay_one_card_thumbnail_url,omitempty"`
}

type EmployeeUpdates struct {
	DGEPayUserID string                `json:"dgepay_user_id,omitempty" gorm:"column:dgepay_user_id"`
	Status       models.EmployeeStatus `json:"status,omitempty"`
	OneID        string                `json:"one_id,omitempty"`
}

type BulkInviteUserPayload struct {
	EmployeeIDs []string `json:"employee_ids" validate:"required,min=1"`
}

type BulkInviteResult struct {
	SuccessCount int                   `json:"success_count"`
	FailedCount  int                   `json:"failed_count"`
	Results      []InviteResult        `json:"results"`
	FailedItems  []BulkInviteErrorItem `json:"failed_items"`
}

type InviteResult struct {
	EmployeeID string `json:"employee_id"`
	Email      string `json:"email"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

type BulkInviteErrorItem struct {
	EmployeeID string `json:"employee_id"`
	Error      string `json:"error"`
}

type EmployeeSearchRequest struct {
	Pagination

	Search      string `query:"search"` // OR search across name, phone, email
	UserName    string `query:"user_name"`
	Designation string `query:"designation"`
	Email       string `query:"email"`
	Phone       string `query:"phone"`
	Status      string `query:"status"`
}

type MerchantSearchRequest struct {
	Pagination
	BIN              string `json:"bin"`
	BusinessName     string `json:"business_name"`
	Email            string `json:"email"`
	PhoneNumber      string `json:"phone_number"`
	Status           string `json:"status"`
	FirstName        string `json:"first_name"`
	LastName         string `json:"last_name"`
	NID              string `json:"nid"`
	Address          string `json:"address"`
	BusinessCategory string `json:"business_category"`
	TINNo            string `json:"tin_no"`
	CreatedFrom      string `json:"created_from"`
	CreatedTo        string `json:"created_to"`
}

type EmployeeInviteCandidate struct {
	ID         string
	MerchantID string
}

type MembershipResponse struct {
	ID          string `json:"id"`
	Designation string `json:"designation"`
	Status      string `json:"status"`
	RefID       string `json:"ref_id"`

	Employee EmployeeDTO `json:"employee"`
	Card     CardDTO     `json:"card"`
}

type EmployeeDTO struct {
	ID       string `json:"id"`
	UserName string `json:"user_name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
}

type CardDTO struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type BulkDownloadCardsRequest struct {
	MembershipIDs []string `json:"membership_ids" validate:"required,min=1,max=1000"`
}

type ChangeMembershipStatusRequest struct {
	Status models.MembershipStatus `json:"status" validate:"required"`
	Remark string                  `json:"remark" validate:"required"`
}

type ChangeMembershipStatusResponse struct {
	ID     string                  `json:"id"`
	Status models.MembershipStatus `json:"status"`
	Remark string                  `json:"remark"`
}

type ChangeEmployeeCardStatusRequest struct {
	Status models.EmployeeCardStatus `json:"status" validate:"required"`
	Remark string                    `json:"remark" validate:"required"`
}

type ChangeEmployeeCardStatusResponse struct {
	ID     string                    `json:"id"`
	Status models.EmployeeCardStatus `json:"status"`
	Remark string                    `json:"remark"`
}

type UploadAvatarResponse struct {
	AvatarURL  string  `json:"avatar_url"`
	FaceMatch  bool    `json:"face_match"`
	Similarity float64 `json:"similarity"`
	Threshold  float64 `json:"threshold"`
}

type UploadEmployeeFileResponse struct {
	URL       string `json:"url"`
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
	Format    string `json:"format"`
}

type EmployeeProfileResponse struct {
	ID           string `json:"id"`
	MembershipID string `json:"membership_id"`
	AvatarURL    string `json:"avatar_url"`
}

type BulkUploadResult struct {
	SuccessCount int               `json:"success_count"`
	FailedCount  int               `json:"failed_count"`
	FailedRows   []BulkUploadError `json:"failed_rows"`
}

type BulkUploadError struct {
	Row    int                    `json:"row"`
	Error  string                 `json:"error"`
	Fields []BulkUploadFieldError `json:"fields,omitempty"`
}

type BulkUploadFieldError struct {
	Field   string `json:"field,omitempty"`
	Value   string `json:"value,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginatedEmployees struct {
	Employees  []models.Employee `json:"employees"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type DownloadCardAsset struct {
	Path string
	Data []byte
}

type UpdateVDSDocumentRequest struct {
	EmployeeID  string   `json:"employee_id,omitempty"`
	EmployeeIDs []string `json:"employee_ids,omitempty"`
}

type UpdateVDSDocumentResult struct {
	EmployeeID string           `json:"employee_id"`
	Success    bool             `json:"success"`
	Data       *GenerateVDSData `json:"data,omitempty"`
	Error      string           `json:"error,omitempty"`
}

type BulkUpdateVDSDocumentResponse struct {
	TotalCount   int                       `json:"total_count"`
	SuccessCount int                       `json:"success_count"`
	FailedCount  int                       `json:"failed_count"`
	Results      []UpdateVDSDocumentResult `json:"results"`
}

type UpdateCardUIRequest struct {
	MembershipID  string   `json:"membership_id,omitempty"`
	MembershipIDs []string `json:"membership_ids,omitempty"`
}

type UpdateCardUIResult struct {
	MembershipID string                `json:"membership_id"`
	Success      bool                  `json:"success"`
	Result       *GenerateCardResponse `json:"result,omitempty"`
	Error        string                `json:"error,omitempty"`
}

type BulkUpdateCardUIResponse struct {
	TotalCount   int                  `json:"total_count"`
	SuccessCount int                  `json:"success_count"`
	FailedCount  int                  `json:"failed_count"`
	Results      []UpdateCardUIResult `json:"results"`
}
