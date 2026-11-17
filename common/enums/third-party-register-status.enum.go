package enums

type RegistrationStatus string

const (
	RegistrationStatusApproved RegistrationStatus = "approved"
	RegistrationStatusPending  RegistrationStatus = "pending"
	RegistrationStatusRejected RegistrationStatus = "rejected"
)
