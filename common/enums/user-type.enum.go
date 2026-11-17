package enums

import "database/sql/driver"

type UserType string

const (
	UserTypeAdmin      UserType = "admin"
	UserTypeMerchant   UserType = "merchant"
)

func (ut UserType) Value() (driver.Value, error) {
	return string(ut), nil
}

func (ut *UserType) Scan(value any) error {
	if value == nil {
		*ut = ""
		return nil
	}
	*ut = UserType(value.(string))
	return nil
}
