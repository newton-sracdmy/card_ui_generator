package constant

import "time"

const version = "1.0.3"
const UserIDKey = "userID"
const UserClaimsKey = "userClaims"
const (
	DG_E_PAY_MONTHLY_TURNOVER        = 100000
	DG_E_PAY_VS_CHALLAN_REGISTRATION = true
	DGEPayDialCode                   = "+88"
)

type DGEPayUserType int

const (
	DGEPayUserTypeCustomer DGEPayUserType = 2
	DGEPayUserTypeMerchant DGEPayUserType = 3
)

const (
	QueryParamMembershipID = "membership_id"
)

// background service
const (
	EmployeeInviteInterval       = 24 * time.Hour
	NatsLockTTL                  = 30 * time.Minute
	YogiOnboardReconcileInterval = 15 * time.Minute
	// Memberships younger than this are left alone; their first onboarding job is probably still queued.
	YogiOnboardReconcileGrace   = 10 * time.Minute
	BulkDownloadJobStatusTTL    = 24 * time.Hour
	BulkDownloadJobStatusBucket = "bulk_download_jobs"
	EmployeeCardCacheBucket     = "employee_card_cache"
	EmployeeCardCacheTTL        = 24 * time.Hour
)
const BackgroundLockBucket = "background_service_locks"

// OneID eKYC
const (
	OneIDEKYCIdentityType           = "nid"
	OneIDEKYCSessionValiditySeconds = 864000 // 10 days
)

// OneID service provider
const (
	OneIDServiceProviderCreatedFrom = "one_card"
)

// oneId vault type
const (
	SyncDocumentType = "Certificate"
)

// dgepay card type
const (
	DgepayCardType = "ONE_CARD"
)

// S3 Upload Folders
const (
	AvatarUploadFolder = "one_id_user"
)

// Designations
const (
	DesignationManager = "Managing Director, Dhaka Bank PLC"
	DesignationFaculty = "Faculty"
	DesignationStudent = "Student"
	DesignationYogi    = "Yogi"
)
