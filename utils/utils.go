package utils

import (
	"context"
	"crypto/rand"
	"employee-card-admin/models"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type JSONMap map[string]interface{}

type QRMetadata map[string]map[string]string

func GenerateFourDigitPIN() (string, error) {
	if os.Getenv("ENVIRONMENT") != "production" {
		return "1234", nil
	}

	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", fmt.Errorf("FAILED_TO_GENERATE_PIN: %w", err)
	}

	return fmt.Sprintf("%04d", n.Int64()), nil
}

func GenerateUniqueCardNumber() (string, error) {
	cardNumber := make([]byte, 12)
	for i := range cardNumber {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", fmt.Errorf("failed to generate random digit: %w", err)
		}
		cardNumber[i] = byte(num.Int64() + '0')
	}
	return string(cardNumber), nil
}

func ConvertJSONMapToQRMetadata(jsonData JSONMap) QRMetadata {
	qrMetadata := make(QRMetadata)
	for key, value := range jsonData {
		if valueMap, ok := value.(map[string]interface{}); ok {
			qrMetadata[key] = make(map[string]string)
			for subKey, subValue := range valueMap {
				if strValue, ok := subValue.(string); ok {
					qrMetadata[key][subKey] = strValue
				}
			}
		} else if strValue, ok := value.(string); ok {
			qrMetadata[key] = map[string]string{"value": strValue}
		}
	}
	return qrMetadata
}

func AddQueryParamToURL(rawURL string, key string, value string, userName string, refID string, oneID string, hasAcount bool) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("INVALID_URL: %w", err)
	}
	query := parsedURL.Query()
	query.Set(key, value)
	query.Set("userName", userName)

	if hasAcount {
		query.Set("ref_id", refID)
		query.Set("one_id", oneID)
	}
	// query.Set("api_key", os.Getenv("ONE_ID_API_KEY"))
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

func ConvertDOBToTimestamp(dob string) (int64, error) {
	dob = strings.TrimSpace(dob)
	formats := []string{"02/01/2006", "2006-01-02", "02-01-2006", "2006/01/02"}
	for _, format := range formats {
		if t, err := time.Parse(format, dob); err == nil {
			return t.Unix(), nil
		}
	}

	return 0, fmt.Errorf("invalid date format: %s", dob)
}

func GetStringFromMetaData(metaData models.JSONMap, key string) string {
	if value, ok := metaData[key].(string); ok {
		return value
	}
	return ""
}

func DownloadFile(ctx context.Context, client *http.Client, cardURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cardURL, nil)
	if err != nil {
		return nil, NewInternalServerError("RESPONSE_MESSAGES_FAILED_TO_BUILD_DOWNLOAD_REQUEST", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, NewInternalServerError("RESPONSE_MESSAGES_FAILED_TO_DOWNLOAD_CARD", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, NewInternalServerError("RESPONSE_MESSAGES_FAILED_TO_DOWNLOAD_CARD", fmt.Errorf("HTTP %d", resp.StatusCode))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, NewInternalServerError("RESPONSE_MESSAGES_FAILED_TO_READ_CARD_DATA", err)
	}

	return data, nil
}

func GetEKYCRedirectURI(
	email string,
	phone string,
	refID string,
	finalRedirectURI string,
) (string, error) {
	if finalRedirectURI == "" {
		return "", fmt.Errorf("final redirect URI is empty")
	}

	if _, err := url.ParseRequestURI(finalRedirectURI); err != nil {
		return "", fmt.Errorf("invalid final redirect URI: %w", err)
	}

	baseURL := os.Getenv("ONE_ID_REGISTER_URI")
	if baseURL == "" {
		return "", fmt.Errorf("ONE_ID_REGISTER_URI is not set")
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse ONE_ID_REGISTER_URI: %w", err)
	}

	parsedURL.RawQuery = fmt.Sprintf(
		"email=%s&phone=%s&ref_id=%s&redirect_to=%s",
		url.QueryEscape(email),
		url.QueryEscape(phone),
		url.QueryEscape(refID),
		url.QueryEscape(finalRedirectURI),
	)

	return parsedURL.String(), nil
}
