package utils

import (
	"encoding/json"
	"fmt"
	"strings"

	"employee-card-admin/dto"
)

type DgepayCardEnvelope struct {
	UserEmpCardsData DgepayCardData `json:"user_emp_cards_data"`
}

type DgepayCardData struct {
	CardNumber   string `json:"card_number"`
	ExpiryDate   any    `json:"expiry_date"`
	IssuedAt     any    `json:"issued_at"`
	QRCode       string `json:"qr_code"`
	NFCDataToken string `json:"nfc_data_token"`

	MerchantInfo struct {
		BusinessName string `json:"business_name"`
	} `json:"merchant_info"`

	UserInfo struct {
		PhoneNumber string `json:"phone_number"`
	} `json:"user_info"`
}

func ExtractDgepayCardData(resp dto.GenericResponse) (DgepayCardData, error) {
	var envelope DgepayCardEnvelope
	raw, err := json.Marshal(resp.Data)
	if err != nil {
		return DgepayCardData{}, fmt.Errorf("marshal dgepay response: %w", err)
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return DgepayCardData{}, fmt.Errorf("parse dgepay response: %w", err)
	}
	if strings.TrimSpace(envelope.UserEmpCardsData.QRCode) == "" {
		return DgepayCardData{}, fmt.Errorf("payment qr code is missing")
	}
	return envelope.UserEmpCardsData, nil
}
