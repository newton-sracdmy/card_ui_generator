package utils

import (
	"html"
	"net/url"
	"os"
	"strings"

	"employee-card-admin/dto"
)

func BuildMerchantResubmissionLink(merchantID string) string {
	portalURL := strings.TrimSpace(os.Getenv("MERCHANT_RESUBMISSION_PORTAL_URL"))
	if portalURL == "" {
		portalURL = strings.TrimSpace(os.Getenv("CARD_MANAGEMENT_LOGIN_URL"))
	}

	if portalURL == "" {
		return "?merchant_id=" + url.QueryEscape(merchantID)
	}

	parsedURL, err := url.Parse(portalURL)
	if err != nil {
		separator := "?"
		if strings.Contains(portalURL, "?") {
			separator = "&"
		}
		return portalURL + separator + "merchant_id=" + url.QueryEscape(merchantID)
	}

	query := parsedURL.Query()
	query.Set("merchant_id", merchantID)
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String()
}

func BuildFaultyMerchantItemsHTML(body dto.ChangeMerchantStatusRequest) string {
	items := []string{}
	add := func(isFaulty bool, label string) {
		if isFaulty {
			items = append(items, "<li>"+html.EscapeString(label)+"</li>")
		}
	}

	add(body.BIN, "BIN")
	add(body.TINNo, "TIN number")
	add(body.BusinessName, "Business name")
	add(body.PhoneNumber, "Phone number")
	add(body.Email, "Email")
	add(body.OneID, "One ID")
	add(body.OneIDCategoryID, "One ID category")
	add(body.Address, "Address")
	add(body.BusinessCategoryID, "Business category")
	add(body.BusinessServiceTypeID, "Business service type")
	add(body.BusinessServiceTypeName, "Business service type name")
	add(body.BusinessTypeID, "Business type")
	add(body.ServiceProviderDescription, "Service provider description")
	add(body.FirstName, "First name")
	add(body.LastName, "Last name")
	add(body.Logo, "Logo")
	add(body.Signature, "Signature")

	if len(body.Documents) > 0 {
		items = append(items, "<li>Documents</li>")
	}
	if len(body.MerchantOneIDServiceIDs) > 0 {
		items = append(items, "<li>One ID services</li>")
	}

	if len(items) == 0 {
		return "<li>Please review the information marked by the admin and resubmit.</li>"
	}

	return strings.Join(items, "")
}
