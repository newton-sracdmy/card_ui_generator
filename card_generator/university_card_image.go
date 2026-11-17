package card_generator

import (
	"bytes"
	"employee-card-admin/assets/fonts"
	"employee-card-admin/dto"
	"employee-card-admin/models"
	"employee-card-admin/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

func BuildUniversityCardImages(
	employee models.Employee,
	membership models.Membership,
	merchant models.Merchant,
	dgepayCard dto.GenericResponse,
	generateVDS dto.GenerateVDSData,
) (UniversityCardImages, error) {

	// Reset Studio cards carry no payment QR, so route them before parsing the dgepay card.
	if isResetStudioCard(merchant.BusinessName, membership.Designation) {
		return BuildResetStudioMemberCardImages(employee, membership, merchant, dgepayCard, generateVDS)
	}

	paymentData, err := extractUniversityDgepayCardData(dgepayCard)
	if err != nil {
		return UniversityCardImages{}, err
	}

	paymentQRStr := strings.TrimSpace(paymentData.QRCode)
	var paymentQRValue string
	if strings.HasPrefix(paymentQRStr, "http://") || strings.HasPrefix(paymentQRStr, "https://") || strings.HasPrefix(paymentQRStr, "data:") {
		paymentQRValue = paymentQRStr
	} else {
		paymentQRImg, err := generatePaymentQRCodeImage(paymentQRStr, 512)
		if err != nil {
			return UniversityCardImages{}, fmt.Errorf("generate payment qr: %w", err)
		}
		var buf bytes.Buffer
		if err := png.Encode(&buf, paymentQRImg); err != nil {
			return UniversityCardImages{}, fmt.Errorf("encode payment qr png: %w", err)
		}
		paymentQRValue = universityPNGDataURI(buf.Bytes())
	}

	branding := resolveUniversityCardBranding(merchant, merchant.Settings.CreatedAt, merchant.Settings.CardExpiryTime)

	data := universityEmployeeCardTemplateData{
		MemberName:    universityFallbackString(employee.UserName, "N/A"),
		Designation: universityFallbackString(
			universityMetaValue(membership.MetaData, "designation", "faculty_designation", "title"),
			universityFallbackString(membership.Designation, "N/A"),
		),
		RegNo:         universityFallbackString(membership.RefID, "N/A"),
		Program:       universityFallbackString(universityMetaValue(membership.MetaData, "program"), "N/A"),
		Department:    universityFallbackString(universityMetaValue(membership.MetaData, "department"), "N/A"),
		OneID:         universityFallbackString(employee.OneID, "N/A"),
		Initials:      universityBuildInitials(universityFallbackString(employee.UserName, "N/A")),
		AvatarURL:     universityFallbackString(employee.AvatarURL, ""),
		BloodGroup:    universityFallbackString(utils.GetStringFromMetaData(membership.MetaData, "blood_group"), "A+"),
		OneIDQRCode:   universityFallbackImage(generateVDS.ExtractedQRImage),
		PaymentQRCode: paymentQRValue,
		IssueDate:     branding.IssueDate,
		ExpiryDate:    branding.ExpiryDate,
		BusinessName: universityFallbackString(
			merchant.BusinessName,
			paymentData.MerchantInfo.BusinessName,
		),
		PhoneNumber:         "+8802-41010626-9",
		OptionalPhoneNumber: "+8802-58157096",
		FaxNumber:           universityFallbackString(branding.Fax, ""),
		WebsiteUrl:          "www.uap-bd.edu",
		Email:               "registrar@uap-bd.edu",
		BrandColor:          branding.BrandColor,
		LOGO:                "UAP_Student_Card_UI/1790673271-UAP-logo.png",
		Address:             universityFallbackString(branding.Address, "N/A"),
		SignatureURL:        "UAP_Student_Card_UI/1790673528-Registrar's.png",
		SignatureLabel:      "Authorised Signature",
		FooterDisclaimer1:   "This ID card is valid for authorized person only.",
		FooterDisclaimer2:   "If lost & found, please contact the above address.",
	}

	if isManagingDirectorDesignation(membership.Designation) {
		front, err := drawManagingDirectorCardFront(data)
		if err != nil {
			return UniversityCardImages{}, err
		}
		back, err := drawManagingDirectorCardBack(data)
		if err != nil {
			return UniversityCardImages{}, err
		}
		qrBytes, err := universityPNGBytesFromDataURI(data.OneIDQRCode)
		if err != nil {
			qrBytes = nil
		}
		return UniversityCardImages{
			Payment:             back,
			OneID:               front,
			OneIDPNGForSyncData: qrBytes,
		}, nil
	}

	if isFacultyDesignation(membership.Designation) {
		front, err := drawFacultyCardFront(data)
		if err != nil {
			return UniversityCardImages{}, err
		}
		back, err := drawFacultyCardBack(data)
		if err != nil {
			return UniversityCardImages{}, err
		}
		qrBytes, err := universityPNGBytesFromDataURI(data.OneIDQRCode)
		if err != nil {
			qrBytes = nil
		}
		return UniversityCardImages{
			Payment:             front,
			OneID:               back,
			OneIDPNGForSyncData: qrBytes,
		}, nil
	}

	if isStudentDesignation(membership.Designation) {
		front, err := drawStudentCardFront(data)
		if err != nil {
			return UniversityCardImages{}, err
		}
		back, err := drawStudentCardBack(data)
		if err != nil {
			return UniversityCardImages{}, err
		}
		qrBytes, err := universityPNGBytesFromDataURI(data.OneIDQRCode)
		if err != nil {
			qrBytes = nil
		}
		return UniversityCardImages{
			Payment:             front,
			OneID:               back,
			OneIDPNGForSyncData: qrBytes,
		}, nil
	}

	if isSaintsClubCard(merchant.BusinessName, membership.Designation) {
		front, err := drawSaintsClubMemberCardFront(data)
		if err != nil {
			return UniversityCardImages{}, err
		}
		back, err := drawSaintsClubMemberCardBack(data)
		if err != nil {
			return UniversityCardImages{}, err
		}
		thumbnail, err := BuildSaintsClubMemberThumbnailImage(data)
		if err != nil {
			thumbnail = nil
		}
		qrBytes, err := universityPNGBytesFromDataURI(data.OneIDQRCode)
		if err != nil {
			qrBytes = nil
		}
		return UniversityCardImages{
			Payment:             front,
			OneID:               back,
			OneIDPNGForSyncData: qrBytes,
			Thumbnail:           thumbnail,
		}, nil
	}

	front, err := drawDefaultUniversityCardFront(data)
	if err != nil {
		return UniversityCardImages{}, err
	}
	back, err := drawDefaultUniversityCardBack(data)
	if err != nil {
		return UniversityCardImages{}, err
	}
	qrBytes, err := universityPNGBytesFromDataURI(data.OneIDQRCode)
	if err != nil {
		qrBytes = nil
	}
	return UniversityCardImages{
		Payment:             front,
		OneID:               back,
		OneIDPNGForSyncData: qrBytes,
	}, nil
}

const (
	staticAddress               = "Rangs FC Square (Level 03,04,10,11), Plot 6/A, Road 32, Gulshan Avenue, Dhaka - 1212, Bangladesh."
	cardFrontW                  = 4050
	cardFrontH                  = 6300
	cardBackW                   = 4050
	cardBackH                   = 6300
	cardWidthPx                 = 4050
	cardHeightPx                = 6300
	universityDefaultBrandColor = "#631FB4"

	bloodIconB64 = "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"
	nfcIconB64   = "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"
)

type UniversityCardImages struct {
	OneID               []byte
	Payment             []byte
	OneIDPNGForSyncData []byte
	Thumbnail           []byte
}

type CardImages = UniversityCardImages
type dgepayCardEnvelope struct {
	UserEmpCardsData dgepayCardData `json:"user_emp_cards_data"`
}

type dgepayCardData struct {
	CardNumber   string `json:"card_number"`
	ExpiryDate   any    `json:"expiry_date"`
	IssuedAt     any    `json:"issued_at"`
	QRCode       string `json:"qr_code"`
	MerchantInfo struct {
		BusinessName string `json:"business_name"`
	} `json:"merchant_info"`
	UserInfo struct {
		PhoneNumber string `json:"phone_number"`
	} `json:"user_info"`
}

type universityEmployeeCardTemplateData struct {
	MemberName          string
	Designation         string
	RegNo               string
	Program             string
	Department          string
	OneID               string
	Initials            string
	AvatarURL           string
	BloodGroup          string
	OneIDQRCode         string
	PaymentQRCode       string
	IssueDate           string
	ExpiryDate          string
	BusinessName        string
	PhoneNumber         string
	OptionalPhoneNumber string
	FaxNumber           string
	WebsiteUrl          string
	Email               string
	BrandColor          string
	LOGO                string
	Address             string
	SignatureLabel      string
	SignatureURL        string
	FooterDisclaimer1   string
	FooterDisclaimer2   string
}

type employeeCardTemplateData = universityEmployeeCardTemplateData

type universityCardBranding struct {
	BrandColor          string
	Address             string
	OptionalPhoneNumber string
	Fax                 string
	WebsiteURL          string
	SignatureURL        string
	IssueDate           string
	ExpiryDate          string
}

type universityDgepayCardEnvelope struct {
	UserEmpCardsData universityDgepayCardData `json:"user_emp_cards_data"`
}

type universityDgepayCardData struct {
	CardNumber   string `json:"card_number"`
	ExpiryDate   any    `json:"expiry_date"`
	IssuedAt     any    `json:"issued_at"`
	QRCode       string `json:"qr_code"`
	MerchantInfo struct {
		BusinessName string `json:"business_name"`
	} `json:"merchant_info"`
	UserInfo struct {
		PhoneNumber string `json:"phone_number"`
	} `json:"user_info"`
}

func extractUniversityDgepayCardData(resp dto.GenericResponse) (universityDgepayCardData, error) {
	var envelope universityDgepayCardEnvelope
	raw, err := json.Marshal(resp.Data)
	if err != nil {
		return universityDgepayCardData{}, fmt.Errorf("marshal dgepay response: %w", err)
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return universityDgepayCardData{}, fmt.Errorf("parse dgepay response: %w", err)
	}
	if strings.TrimSpace(envelope.UserEmpCardsData.QRCode) == "" {
		return universityDgepayCardData{}, fmt.Errorf("payment qr code is missing")
	}
	return envelope.UserEmpCardsData, nil
}

func resolveUniversityCardBranding(merchant models.Merchant, cardIssued, cardExpiry any) universityCardBranding {
	branding := universityCardBranding{
		BrandColor: universityDefaultBrandColor,
		Address:    "N/A",
		IssueDate:  universityFormatCardDate(cardIssued),
		ExpiryDate: universityFormatCardDate(cardExpiry),
	}
	if merchant.Settings != nil {
		s := merchant.Settings
		if c := strings.TrimSpace(s.BrandColor); c != "" {
			branding.BrandColor = c
		}
		if a := strings.TrimSpace(s.Address); a != "" {
			branding.Address = a
		}
		branding.SignatureURL = strings.TrimSpace(s.Signature)
		branding.OptionalPhoneNumber = universityOptionalStringValue(s.OptionalPhoneNumber)
		branding.Fax = universityOptionalStringValue(s.Fax)
		branding.WebsiteURL = universityOptionalStringValue(s.WebsiteURL)
		if !s.CreatedAt.IsZero() {
			branding.IssueDate = s.CreatedAt.Format("02 January 2006")
		}
		if !s.CardExpiryTime.IsZero() {
			branding.ExpiryDate = s.CardExpiryTime.Format("02 January 2006")
		}
	}
	return branding
}

func universityOptionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func universityMetaValue(metaData models.JSONMap, keys ...string) string {
	for _, key := range keys {
		if v := strings.TrimSpace(utils.GetStringFromMetaData(metaData, key)); v != "" {
			return v
		}
	}
	return ""
}

func universityFormatCardDate(value any) string {
	switch v := value.(type) {
	case nil:
		return "N/A"
	case float64:
		return time.Unix(int64(v), 0).Format("02 January 2006")
	case int64:
		return time.Unix(v, 0).Format("02 January 2006")
	case int:
		return time.Unix(int64(v), 0).Format("02 January 2006")
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return "N/A"
		}
		return time.Unix(n, 0).Format("02 January 2006")
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return "N/A"
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t.Format("02 January 2006")
		}
		return s
	default:
		return "N/A"
	}
}

func universityFallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func universityPNGDataURI(pngBytes []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
}

func universityFallbackImage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNDAiIGhlaWdodD0iMjQwIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGRvbWluYW50LWJhc2VsaW5lPSJtaWRkbGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZpbGw9IiM4ODgiIGZvbnQtc2l6ZT0iMTYiPk5vIFFSPC90ZXh0Pjwvc3ZnPg=="
	}
	return value
}

func universityBuildInitials(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "NA"
	}
	var b strings.Builder
	for _, p := range parts {
		if p != "" {
			b.WriteString(strings.ToUpper(p[:1]))
		}
		if b.Len() == 2 {
			break
		}
	}
	if b.Len() == 0 {
		return "NA"
	}
	return b.String()
}

func universityLogoAsDataURI(merchant models.Merchant) string {
	if merchant.Settings == nil {
		return ""
	}
	rawURL := strings.TrimSpace(merchant.Settings.Logo)
	if rawURL == "" {
		return ""
	}
	if dataURI := universityURLToDataURI(rawURL); dataURI != "" {
		return dataURI
	}
	return rawURL
}

func universityURLToDataURI(rawURL string) string {
	resp, err := http.Get(rawURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return ""
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return ""
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func drawOneIDCard(data employeeCardTemplateData) ([]byte, error) {
	w, h := cardFrontW, cardFrontH
	dc := gg.NewContext(w, h)
	drawCardBackground(dc, data.BrandColor)

	panelX := 0.109 * float64(w)
	panelY := 0.0582 * float64(h)
	panelW := 0.7843 * float64(w)
	panelH := 0.1633 * float64(h)
	drawRoundedPanel(dc, panelX, panelY, panelW, panelH, 8, color.White)

	drawAvatarImage(dc, panelX+72, panelY+panelH-40, 62, data.AvatarURL, data.Initials)

	drawText(dc, "UNIVERSITY OF", 0.49*float64(w), panelY+panelH/2, mustFont(fonts.AptosBold, 12), hexColor("#ab2753ff"), 0, 0.5)
	drawText(dc, "ASIA PACIFIC", 0.49*float64(w), panelY+panelH/1.5, mustFont(fonts.AptosBold, 12), hexColor("#01003A"), 0, 0.5)

	drawText(dc, strings.ToUpper(data.MemberName), 0.1061*float64(w), 0.2795*float64(h),
		mustFont(fonts.AptosBold, 18), color.White, 0, 0.5)
	drawText(dc, data.Program, 0.1061*float64(w), 0.3124*float64(h),
		mustFont(fonts.AptosBold, 18), color.White, 0, 0.5)

	infoX := 0.1052 * float64(w)
	infoY := 0.388 * float64(h)
	infoW := 0.7899 * float64(w)
	infoH := 0.5508 * float64(h)
	drawRoundedPanel(dc, infoX, infoY, infoW, infoH, 10, color.White)

	drawText(dc, "Personal Information", 0.5*float64(w), 0.4095*float64(h)+9,
		mustFont(fonts.AptosBold, 12.66), hexColor(data.BrandColor), 0.5, 0.5)

	qrX := 0.1437 * float64(w)
	qrY := 0.4571 * float64(h)
	qrW := 0.7083 * float64(w)
	qrH := 0.445 * float64(h)
	qrImg, err := generateVDSQRCodeImage(data.OneIDQRCode, 1200)
	if err == nil {
		drawImageContain(dc, qrImg, qrX, qrY, qrW, qrH)
	}

	bgX := 0.80 * float64(w)
	bgY := 0.42 * float64(h)
	drawBloodIcon(dc, bgX, bgY, 44, color.White)
	drawText(dc, data.BloodGroup, bgX, bgY, mustFont(fonts.AptosBold, 14), color.White, 0.5, 0.5)

	dc.Push()
	dc.Translate(infoX+infoW+14, 0.65*float64(h))
	dc.Rotate(-math.Pi / 2)
	drawText(dc, "ID - "+data.OneID, 0, 0, mustFont(fonts.AptosRegular, 20), color.White, 0.5, 0.5)
	dc.Pop()
	dc.ResetClip()

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode one-id image: %w", err)
	}
	return out.Bytes(), nil
}

func drawPaymentCard(data employeeCardTemplateData) ([]byte, error) {
	const w, h = cardBackW, cardBackH
	dc := gg.NewContext(w, h)
	drawCardBackground(dc, data.BrandColor)

	pnlX := 0.1105 * float64(w)
	pnlY := 0.0909 * float64(h)
	pnlW := 0.7956 * float64(w)
	pnlH := 0.3295 * float64(h)
	drawRoundedPanel(dc, pnlX, pnlY, pnlW, pnlH, 8, color.White)

	qrImg, err := generatePaymentQRCodeImage(data.PaymentQRCode, 1000)
	if err != nil {
		return nil, fmt.Errorf("generate payment qr: %w", err)
	}
	qrX := 0.138 * float64(w)
	qrY := 0.1107 * float64(h)
	qrW := 0.4672 * float64(w)
	qrH := 0.2937 * float64(h)
	drawImageContain(dc, qrImg, qrX, qrY, qrW, qrH)

	drawText(dc, "Payment", 0.6575*float64(w), 0.0941*float64(h)+12,
		mustFont(fonts.AptosBold, 24), hexColor(data.BrandColor), 0, 0.5)

	nfcX := 0.7627 * float64(w)
	nfcY := 0.2648 * float64(h)
	drawNFCIcon(dc, nfcX, nfcY, 70, hexColor(data.BrandColor))

	labelX := 0.1035 * float64(w)
	colonX := 0.3507 * float64(w)
	valueX := 0.411 * float64(w)
	issueY := 0.4308 * float64(h)
	expiryY := 0.461 * float64(h)

	drawText(dc, "Issue Date", labelX, issueY, mustFont(fonts.AptosRegular, 12.6), hexColor("#C7A4FF"), 0, 0.5)
	drawText(dc, ":", colonX, issueY, mustFont(fonts.AptosRegular, 12.6), color.White, 0, 0.5)
	drawText(dc, data.IssueDate, valueX, issueY, mustFont(fonts.AptosRegular, 12.6), color.White, 0, 0.5)

	drawText(dc, "Expiry Date", labelX, expiryY, mustFont(fonts.AptosRegular, 12.6), hexColor("#C7A4FF"), 0, 0.5)
	drawText(dc, ":", colonX, expiryY, mustFont(fonts.AptosRegular, 12.6), color.White, 0, 0.5)
	drawText(dc, data.ExpiryDate, valueX, expiryY, mustFont(fonts.AptosRegular, 12.6), color.White, 0, 0.5)

	drawText(dc, data.BusinessName, 0.1149*float64(w), 0.5548*float64(h),
		mustFont(fonts.AptosBold, 15.1), hexColor("#C7A4FF"), 0, 0.5)

	drawText(dc, data.PhoneNumber, 0.1678*float64(w), 0.6014*float64(h),
		mustFont(fonts.AptosBold, 10.8), color.White, 0, 0.5)

	dc.SetFontFace(mustFont(fonts.AptosRegular, 8.9))
	dc.SetColor(color.White)
	dc.DrawStringWrapped(data.Address, 0.167*float64(w), 0.6382*float64(h), 0, 0, 260, 1.4, gg.AlignLeft)

	sigY := 0.7600 * float64(h)
	sigH := 0.0945 * float64(h)
	drawRoundedPanel(dc, 0, sigY, float64(w), sigH, 0, color.White)

	lineY := sigY + sigH*0.65
	dc.SetColor(hexColor("#444444"))
	dc.SetLineWidth(1)
	dc.DrawLine(20, lineY, 140, lineY)
	dc.Stroke()

	drawText(dc, data.SignatureLabel, 0.5409*float64(w), 0.807*float64(h),
		mustFont(fonts.AptosBold, 9.3), hexColor("#0B0B0B"), 0, 0.5)

	dc.SetFontFace(mustFont(fonts.AptosBold, 12))
	dc.SetColor(color.White)
	dc.DrawStringWrapped(
		data.FooterDisclaimer1+"\n"+data.FooterDisclaimer2,
		float64(w)/2, 0.8543*float64(h), 0.5, 0, float64(w)*0.75, 1.5, gg.AlignCenter,
	)

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode payment image: %w", err)
	}
	return out.Bytes(), nil
}

func drawCardBackground(dc *gg.Context, brandColor string) {
	w := dc.Width()
	h := dc.Height()
	dc.SetColor(hexColor(brandColor))
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()

	drawDotStrip(dc, w, 0, 28)
	drawDotStrip(dc, w, h-28, 28)
}

func drawDotStrip(dc *gg.Context, cardW, startY, stripH int) {
	dc.SetColor(withAlpha(hexColor("#FFFFFF"), 0.25))
	spacing := int(float64(cardW) * 0.0123)
	if spacing < 5 {
		spacing = 5
	}
	for x := spacing; x < cardW; x += spacing {
		for y := startY + 2; y < startY+stripH; y += spacing {
			dc.DrawCircle(float64(x), float64(y), 1.0)
			dc.Fill()
		}
	}
}

func drawRoundedPanel(dc *gg.Context, x, y, w, h, r float64, fill color.Color) {
	dc.SetColor(fill)
	dc.DrawRoundedRectangle(x, y, w, h, r)
	dc.Fill()
}

func drawCircle(dc *gg.Context, x, y, radius float64, fill color.Color) {
	dc.SetColor(fill)
	dc.DrawCircle(x, y, radius)
	dc.Fill()
}

func scaleImageNearest(src image.Image, width, height int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW == 0 || srcH == 0 {
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	scaleX := float64(srcW) / float64(width)
	scaleY := float64(srcH) / float64(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcX := int(float64(x) * scaleX)
			srcY := int(float64(y) * scaleY)
			if srcX >= srcW {
				srcX = srcW - 1
			}
			if srcY >= srcH {
				srcY = srcH - 1
			}
			dst.Set(x, y, src.At(srcX+bounds.Min.X, srcY+bounds.Min.Y))
		}
	}
	return dst
}

func generateVDSQRCodeImage(data string, size int) (image.Image, error) {
	qr, err := qrcode.New(data, qrcode.High)
	if err != nil {
		return nil, err
	}
	qr.DisableBorder = true
	qr.ForegroundColor = color.Black
	qr.BackgroundColor = color.White
	return qr.Image(size), nil
}

func generatePaymentQRCodeImage(data string, size int) (image.Image, error) {
	qr, err := qrcode.New(data, qrcode.High)
	if err != nil {
		return nil, err
	}
	qr.DisableBorder = true
	qr.ForegroundColor = hexColor("#0F4C9C")
	qr.BackgroundColor = color.White
	return qr.Image(size), nil
}

func ensureBlackQRCodeImage(img image.Image) image.Image {
	if img == nil {
		return nil
	}
	bounds := img.Bounds()
	dst := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			if a > 0 {
				luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
				if luminance < 0.5 {
					dst.Set(x, y, color.RGBA{R: 0, G: 0, B: 0, A: uint8(a >> 8)})
					continue
				}
			}
			dst.Set(x, y, c)
		}
	}
	return dst
}

func loadEmbeddedPNG(b64Data string) image.Image {
	raw, err := base64.StdEncoding.DecodeString(b64Data)
	if err != nil {
		return nil
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil
	}
	return img
}

func loadImageFromURIOrURL(value string) (image.Image, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty image source")
	}

	// 1. Check in-memory cache
	if cached, ok := cardAssetCache.Load(value); ok && cached != nil {
		if img, ok := cached.(image.Image); ok {
			return img, nil
		}
	}

	var loadedImg image.Image
	var loadErr error

	switch {
	case strings.HasPrefix(value, "data:"):
		loadedImg, loadErr = decodeDataURIToImage(value)

	case isS3Source(value):
		if data, err := os.ReadFile(value); err == nil {
			if img, _, errDec := image.Decode(bytes.NewReader(data)); errDec == nil {
				loadedImg = img
				loadErr = nil
				break
			}
		}
		loadedImg, loadErr = downloadImageFromS3(value)
		if loadErr != nil && !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
			// Fall back to embedded base64 if S3 download fails
			if emb := loadEmbeddedPNG(value); emb != nil {
				loadedImg = emb
				loadErr = nil
			}
		}

	case strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://"):
		reqURL := value
		if strings.Contains(reqURL, "res.cloudinary.com") && strings.Contains(reqURL, "/upload/") && !strings.Contains(reqURL, "/f_png/") {
			reqURL = strings.Replace(reqURL, "/upload/", "/upload/w_3000,f_png/", 1)
		}
		resp, err := http.Get(reqURL)
		if err == nil {
			defer resp.Body.Close()
			img, _, errDec := image.Decode(resp.Body)
			if errDec == nil {
				loadedImg = img
			}
		}
		if loadedImg == nil && reqURL != value {
			resp2, err2 := http.Get(value)
			if err2 == nil {
				defer resp2.Body.Close()
				img2, _, err2Dec := image.Decode(resp2.Body)
				if err2Dec == nil {
					loadedImg = img2
				}
			}
		}
		if loadedImg == nil {
			loadErr = fmt.Errorf("unable to decode image from URL %s", value)
		}

	default:
		// Try S3 first for any non-URL string
		loadedImg, loadErr = downloadImageFromS3(value)
		if loadErr != nil || loadedImg == nil {
			if emb := loadEmbeddedPNG(value); emb != nil {
				loadedImg = emb
				loadErr = nil
			} else {
				loadErr = fmt.Errorf("unable to decode image source: %s", value)
			}
		}
	}

	if loadErr != nil || loadedImg == nil {
		if loadErr == nil {
			loadErr = fmt.Errorf("empty image result for: %s", value)
		}
		return nil, loadErr
	}

	// Cache successful decoded image
	cardAssetCache.Store(value, loadedImg)
	return loadedImg, nil
}


func loadImageOrGenerateVDSQRCode(value string, size int) (image.Image, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty VDS qr value")
	}
	if strings.HasPrefix(value, "data:") {
		img, err := decodeDataURIToImage(value)
		if err == nil {
			return ensureBlackQRCodeImage(img), nil
		}
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		img, err := loadImageFromURIOrURL(value)
		if err == nil && img != nil {
			return ensureBlackQRCodeImage(img), nil
		}
	}
	if img := loadEmbeddedPNG(value); img != nil {
		return ensureBlackQRCodeImage(img), nil
	}
	return generateVDSQRCodeImage(value, size)
}

func loadImageOrGeneratePaymentQRCode(value string, size int) (image.Image, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("empty payment qr value")
	}
	var img image.Image
	var err error
	if strings.HasPrefix(value, "data:") {
		img, err = decodeDataURIToImage(value)
	} else if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		img, err = loadImageFromURIOrURL(value)
	} else if embedded := loadEmbeddedPNG(value); embedded != nil {
		img = embedded
	} else {
		img, err = generatePaymentQRCodeImage(value, size)
	}
	if err != nil || img == nil {
		return nil, fmt.Errorf("failed to load/generate payment QR: %v", err)
	}
	return img, nil
}

func drawInitialsBadge(dc *gg.Context, x, y, radius float64, initials string) {
	drawCircle(dc, x, y, radius, hexColor("#ede9fe"))
	fontSize := radius * 0.75
	drawText(dc, fallbackString(initials, "NA"), x, y, mustFont(fonts.AptosBold, fontSize), hexColor("#6d28d9"), 0.5, 0.5)
}

func drawText(dc *gg.Context, text string, x, y float64, face font.Face, clr color.Color, ax, ay float64) {
	dc.SetFontFace(face)
	dc.SetColor(clr)
	dc.DrawStringAnchored(text, x, y, ax, ay)
}

func drawImageContain(dc *gg.Context, img image.Image, x, y, w, h float64) {
	if img == nil {
		return
	}

	bounds := img.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())
	if srcW == 0 || srcH == 0 {
		return
	}

	scale := minFloat(w/srcW, h/srcH)
	drawW := srcW * scale
	drawH := srcH * scale
	posX := x + (w-drawW)/2
	posY := y + (h-drawH)/2

	targetW := int(math.Round(drawW))
	targetH := int(math.Round(drawH))
	if targetW <= 0 {
		targetW = 1
	}
	if targetH <= 0 {
		targetH = 1
	}
	scaledImg := scaleImageHighQuality(img, targetW, targetH)

	dc.Push()
	dc.Translate(posX, posY)
	dc.DrawImage(scaledImg, 0, 0)
	dc.Pop()
	dc.ResetClip()
}

func renderTintedSVGIcon(svgPath string, sourceFill string, fillColor string) (image.Image, error) {
	data, err := os.ReadFile(svgPath)
	if err != nil {
		return nil, fmt.Errorf("read svg icon source: %w", err)
	}

	src := strings.ReplaceAll(string(data), sourceFill, fillColor)

	svgFile, err := os.CreateTemp("", "icon-*.svg")
	if err != nil {
		return nil, fmt.Errorf("create temp svg: %w", err)
	}
	defer os.Remove(svgFile.Name())
	if _, err := svgFile.WriteString(src); err != nil {
		svgFile.Close()
		return nil, fmt.Errorf("write temp svg: %w", err)
	}
	if err := svgFile.Close(); err != nil {
		return nil, fmt.Errorf("close temp svg: %w", err)
	}

	pngFile, err := os.CreateTemp("", "icon-*.png")
	if err != nil {
		return nil, fmt.Errorf("create temp png: %w", err)
	}
	pngPath := pngFile.Name()
	pngFile.Close()
	defer os.Remove(pngPath)

	cmd := exec.Command("convert", svgFile.Name(), "-background", "none", "-alpha", "on", "-transparent", "white", "-trim", "+repage", pngPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("convert svg icon: %w: %s", err, string(output))
	}

	pngBytes, err := os.ReadFile(pngPath)
	if err != nil {
		return nil, fmt.Errorf("read rendered png: %w", err)
	}

	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("decode rendered png: %w", err)
	}
	return img, nil
}

func renderBloodIcon(fillColor string) (image.Image, error) {
	if img := loadEmbeddedPNG(bloodIconB64); img != nil {
		return img, nil
	}
	return renderTintedSVGIcon(filepath.Join(".", "blood_icon.svg"), "#631fb4", fillColor)
}

func renderNFCIcon(fillColor string) (image.Image, error) {
	if img := loadEmbeddedPNG(nfcIconB64); img != nil {
		return img, nil
	}
	return renderTintedSVGIcon(filepath.Join(".", "nfc_icon.svg"), "#5e1eb2", fillColor)
}

func drawBloodIcon(dc *gg.Context, cx, cy, size float64, fill color.Color) {
	r, g, b, a := fill.RGBA()
	fillHex := fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
	img, err := renderBloodIcon(fillHex)
	if err != nil {
		return
	}
	drawImageContain(dc, img, cx-size/2, cy-size/2, size, size)
	_ = a
}

func drawNFCIcon(dc *gg.Context, cx, cy, size float64, fill color.Color) {
	r, g, b, a := fill.RGBA()
	fillHex := fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
	img, err := renderNFCIcon(fillHex)
	if err != nil {
		return
	}
	drawImageContain(dc, img, cx-size/2, cy-size/2, size, size)
	_ = a
}

func decodeDataURIToImage(dataURI string) (image.Image, error) {
	parts := strings.SplitN(strings.TrimSpace(dataURI), ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid image data")
	}

	imgBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode image base64: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(imgBytes))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	return img, nil
}

func universityPNGBytesFromDataURI(dataURI string) ([]byte, error) {
	parts := strings.SplitN(strings.TrimSpace(dataURI), ",", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid PNG data")
	}
	return base64.StdEncoding.DecodeString(parts[1])
}

func pngDataURI(pngBytes []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
}

func fallbackImage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNDAiIGhlaWdodD0iMjQwIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGRvbWluYW50LWJhc2VsaW5lPSJtaWRkbGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZpbGw9IiM4ODgiIGZvbnQtc2l6ZT0iMTYiPk5vIFFSPC90ZXh0Pjwvc3ZnPg=="
	}
	return value
}

func buildInitials(name string) string {
	parts := strings.Fields(strings.TrimSpace(name))
	if len(parts) == 0 {
		return "NA"
	}

	var initials strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		initials.WriteString(strings.ToUpper(part[:1]))
		if initials.Len() == 2 {
			break
		}
	}

	if initials.Len() == 0 {
		return "NA"
	}

	return initials.String()
}

func formatCardDate(value any) string {
	switch v := value.(type) {
	case nil:
		return "N/A"
	case float64:
		return time.Unix(int64(v), 0).Format("02 January 2006")
	case int64:
		return time.Unix(v, 0).Format("02 January 2006")
	case int:
		return time.Unix(int64(v), 0).Format("02 January 2006")
	case json.Number:
		n, err := v.Int64()
		if err != nil {
			return "N/A"
		}
		return time.Unix(n, 0).Format("02 January 2006")
	case string:
		value := strings.TrimSpace(v)
		if value == "" {
			return "N/A"
		}
		if t, err := time.Parse(time.RFC3339, value); err == nil {
			return t.Format("02 January 2006")
		}
		return value
	default:
		return "N/A"
	}
}

func fallbackString(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func mustFont(ttf []byte, size float64) font.Face {
	if len(ttf) == 0 {
		panic("empty font data provided to mustFont")
	}

	parsed, err := opentype.Parse(ttf)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		panic(err)
	}

	return face
}

func hexColor(value string) color.Color {
	value = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(value) != 6 {
		return color.Black
	}

	var r, g, b uint8
	_, _ = fmt.Sscanf(value, "%02x%02x%02x", &r, &g, &b)
	return color.RGBA{R: r, G: g, B: b, A: 255}
}

func withAlpha(c color.Color, alpha float64) color.Color {
	r, g, b, _ := c.RGBA()
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(maxFloat(0, minFloat(1, alpha)) * 255),
	}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func drawAvatarImage(dc *gg.Context, x, y, radius float64, badgeURL string, initials string) {
	if badgeURL != "" {
		resp, err := http.Get(badgeURL)
		if err == nil {
			defer resp.Body.Close()

			img, _, err := image.Decode(resp.Body)
			if err == nil {
				bounds := img.Bounds()
				imgW := float64(bounds.Dx())
				imgH := float64(bounds.Dy())

				diameter := radius * 2

				// Contain: show the full image without cropping
				scale := math.Min(
					diameter/imgW,
					diameter/imgH,
				)

				scaledW := imgW * scale
				scaledH := imgH * scale

				// Detect background color from left & right edge middle
				left := img.At(0, bounds.Dy()/2)
				right := img.At(bounds.Dx()-1, bounds.Dy()/2)

				r1, g1, b1, a1 := left.RGBA()
				r2, g2, b2, a2 := right.RGBA()

				bg := color.RGBA{
					R: uint8(((r1 + r2) / 2) >> 8),
					G: uint8(((g1 + g2) / 2) >> 8),
					B: uint8(((b1 + b2) / 2) >> 8),
					A: uint8(((a1 + a2) / 2) >> 8),
				}

				dc.Push()

				// Clip to circle
				dc.DrawCircle(x, y, radius)
				dc.Clip()

				// Fill background using detected image background colour
				dc.SetColor(bg)
				dc.DrawCircle(x, y, radius)
				dc.Fill()

				// Draw image centred
				dc.Push()
				dc.Translate(
					x-scaledW/2,
					y-scaledH/2,
				)
				dc.Scale(scale, scale)
				dc.DrawImage(img, 0, 0)
				dc.Pop()

				dc.Pop()
				dc.ResetClip()

				// Draw border
				dc.DrawCircle(x, y, radius)
				dc.SetColor(hexColor("#ede9fe"))
				dc.SetLineWidth(4)
				dc.Stroke()

				return
			}
		}
	}

	drawInitialsBadge(dc, x, y, radius, initials)
}

func drawDefaultUniversityCardFront(data employeeCardTemplateData) ([]byte, error) {
	w, h := cardFrontW, cardFrontH
	dc := gg.NewContext(w, h)

	dc.SetColor(hexColor(data.BrandColor))
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()

	drawDotStrip(dc, w, h-280, 280)

	panelX, panelY, panelW, panelH := 200.0, 300.0, float64(w)-400.0, 1000.0
	drawRoundedPanel(dc, panelX, panelY, panelW, panelH+90, 120, color.White)

	avatarR := 590.0
	avatarCenterX := panelX + 740
	drawAvatarImage(dc, avatarCenterX, panelY+panelH/2+230, avatarR, data.AvatarURL, data.Initials)

	avatarRightEdge := avatarCenterX + avatarR
	remainingStart := avatarRightEdge
	remainingEnd := panelX + panelW
	remainingWidth := remainingEnd - remainingStart
	remainingCenterX := remainingStart + remainingWidth/2

	logoSize := 530.0
	logoX := remainingCenterX - logoSize/2
	logoY := panelY + 130

	logoImg, err := loadImageFromURIOrURL(data.LOGO)
	if err == nil && logoImg != nil {
		drawImageContain(dc, logoImg, logoX, logoY, logoSize, logoSize)
	}

	nameFont := mustFont(fonts.AptosRegular, 130)
	dc.SetFontFace(nameFont)
	nameMaxWidth := remainingWidth - 80
	nameLines := dc.WordWrap(data.BusinessName, nameMaxWidth)
	nameY := logoY + logoSize + 110
	for i, line := range nameLines {
		drawText(dc, line, remainingCenterX, nameY+float64(i*160), nameFont, hexColor("#000000"), 0.5, 0.5)
	}

	drawText(dc, strings.ToUpper(data.MemberName), 200.0, 1800.0, mustFont(fonts.AptosBold, 160), color.White, 0, 0.5)

	bloodImg, err := renderBloodIcon("#ff0000")
	if err == nil {
		drawImageContain(dc, bloodImg, float64(w)-670, 1550, 500, 500)
	}
	drawText(dc, data.BloodGroup, float64(w)-300, 1650, mustFont(fonts.AptosBold, 120), color.White, 0.5, 0.5)

	labelX := 200.0
	valueX := 1200.0

	y := 2100.0
	drawText(dc, "Reg No:", labelX, y, mustFont(fonts.AptosBold, 120), color.White, 0, 0.5)
	drawText(dc, data.RegNo, valueX, y, mustFont(fonts.AptosRegular, 120), color.White, 0, 0.5)

	y += 240.0
	drawText(dc, "Program:", labelX, y, mustFont(fonts.AptosBold, 120), color.White, 0, 0.5)
	dc.SetFontFace(mustFont(fonts.AptosRegular, 120))
	dc.SetColor(color.White)
	lines := dc.WordWrap(data.Program, float64(w)-valueX-200)
	for i, line := range lines {
		drawText(dc, line, valueX, y+float64(i*180), mustFont(fonts.AptosRegular, 120), color.White, 0, 0.5)
	}
	y += float64(len(lines)*180) + 60.0

	drawText(dc, "Department:", labelX, y, mustFont(fonts.AptosBold, 120), color.White, 0, 0.5)
	lines2 := dc.WordWrap(data.Department, float64(w)-valueX-200)
	for i, line := range lines2 {
		drawText(dc, line, valueX, y+float64(i*180), mustFont(fonts.AptosRegular, 120), color.White, 0, 0.5)
	}

	qrPanelY := 3300.0
	qrPanelH := 2000.0
	drawRoundedPanel(dc, 200, qrPanelY, float64(w)-400, qrPanelH, 120, color.White)

	paymentQRImg, err := loadImageOrGeneratePaymentQRCode(data.PaymentQRCode, 1200)
	if err == nil && paymentQRImg != nil {
		drawImageContain(dc, paymentQRImg, 300, qrPanelY+100, 1800, 1800)
	}

	drawText(dc, "Payment", 2900, qrPanelY+500, mustFont(fonts.AptosBold, 160), hexColor(data.BrandColor), 0.5, 0.5)

	nfcImg, err := renderNFCIcon("#ff0000")
	if err == nil {
		drawImageContain(dc, nfcImg, 2550, qrPanelY+800, 700, 700)
	}

	dateY := qrPanelY + qrPanelH + 300
	drawText(dc, "Issue Date:", labelX, dateY, mustFont(fonts.AptosBold, 120), color.White, 0, 0.5)
	drawText(dc, data.IssueDate, valueX, dateY, mustFont(fonts.AptosRegular, 120), color.White, 0, 0.5)

	dateY += 200
	drawText(dc, "Expiry Date:", labelX, dateY, mustFont(fonts.AptosBold, 120), color.White, 0, 0.5)
	drawText(dc, data.ExpiryDate, valueX, dateY, mustFont(fonts.AptosRegular, 120), color.White, 0, 0.5)

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode front image: %w", err)
	}
	return out.Bytes(), nil
}

func drawDefaultUniversityCardBack(data employeeCardTemplateData) ([]byte, error) {
	w, h := cardBackW, cardBackH
	dc := gg.NewContext(w, h)

	dc.SetColor(color.White)
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()

	dc.SetColor(hexColor(data.BrandColor))
	dc.DrawRectangle(0, 0, float64(w), 750)
	dc.Fill()
	drawText(dc, "Personal Information", float64(w)/2, 375, mustFont(fonts.AptosBold, 180), color.White, 0.5, 0.5)

	oneIDQRImg, err := loadImageOrGenerateVDSQRCode(data.OneIDQRCode, 3160)
	if err == nil && oneIDQRImg != nil {
		drawImageContain(dc, oneIDQRImg, 400, 900, 3160, 3160)
	}

	dc.Push()
	dc.Translate(3800, 2480)
	dc.Rotate(-math.Pi / 2)
	drawText(dc, "Scan with the AuthnTick App", 0, 0, mustFont(fonts.AptosRegular, 100), hexColor("#000000"), 0.5, 0.5)
	dc.Pop()

	textBlockMaxWidth := float64(w) - 400
	drawWrappedCentered := func(text string, startY float64, font font.Face, lineHeight float64) float64 {
		if strings.TrimSpace(text) == "" {
			return startY
		}
		dc.SetFontFace(font)
		lines := dc.WordWrap(text, textBlockMaxWidth)
		for i, line := range lines {
			drawText(dc, line, float64(w)/2, startY+float64(i)*lineHeight, font, hexColor("#000000"), 0.5, 0.5)
		}
		return startY + float64(len(lines))*lineHeight
	}

	textY := 4200.0
	boldFont := mustFont(fonts.AptosBold, 160)
	regularFont := mustFont(fonts.AptosRegular, 110)

	textY = drawWrappedCentered(data.BusinessName, textY, boldFont, 200) + 20

	textY = drawWrappedCentered(data.Address, textY, regularFont, 150) + 30

	textY = drawWrappedCentered(strings.TrimSpace(data.PhoneNumber+" "+data.OptionalPhoneNumber), textY, regularFont, 150) + 30

	textY = drawWrappedCentered("Email: "+data.Email, textY, regularFont, 150) + 30

	if strings.TrimSpace(data.WebsiteUrl) != "" {
		textY = drawWrappedCentered("Website: "+data.WebsiteUrl, textY, regularFont, 150) + 30
	}

	sigY := math.Max(5150.0, textY+150)
	dc.SetColor(hexColor("#E6E6EB"))
	dc.DrawRectangle(0, sigY, float64(w), 600)
	dc.Fill()

	dc.SetColor(hexColor("#000000"))
	dc.SetLineWidth(20)
	if strings.TrimSpace(data.SignatureURL) != "" {
		sigImg, errSig := loadImageFromURIOrURL(data.SignatureURL)
		if errSig == nil && sigImg != nil {
			drawImageContain(dc, sigImg, 200, sigY+50, 2200, 500)
		} else {
			imgScribble, errScribble := decodeDataURIToImage("data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNDAiIGhlaWdodD0iMTAwIj48cGF0aCBkPSJNMTAgNTBRMzAgMTAgNTAgNTBUNTAgNTBUMTAwIDEwVDEwMCA1MCIgc3Ryb2tlPSJibGFjayIgZmlsbD0ibm9uZSIgc3Ryb2tlLXdpZHRoPSIyIi8+PC9zdmc+")
			if errScribble == nil && imgScribble != nil {
				drawImageContain(dc, imgScribble, 200, sigY, 1500, 600)
			} else {
				dc.MoveTo(1200, sigY+400)
				for i := 40; i < 200; i += 10 {
					dc.QuadraticTo(float64(i+5+800), sigY+100, float64(i+10+800), sigY+400)
				}
				dc.Stroke()
			}
		}
	} else {
		imgScribble, errScribble := decodeDataURIToImage("data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNDAiIGhlaWdodD0iMTAwIj48cGF0aCBkPSJNMTAgNTBRMzAgMTAgNTAgNTBUNTAgNTBUMTAwIDEwVDEwMCA1MCIgc3Ryb2tlPSJibGFjayIgZmlsbD0ibm9uZSIgc3Ryb2tlLXdpZHRoPSIyIi8+PC9zdmc+")
		if errScribble == nil && imgScribble != nil {
			drawImageContain(dc, imgScribble, 200, sigY, 1500, 600)
		} else {
			dc.MoveTo(1200, sigY+400)
			for i := 40; i < 200; i += 10 {
				dc.QuadraticTo(float64(i+5+800), sigY+100, float64(i+10+800), sigY+400)
			}
			dc.Stroke()
		}
	}
	drawText(dc, "Authorised Signature", 3200, sigY+300, mustFont(fonts.AptosBold, 100), hexColor("#000000"), 0.5, 0.5)

	disclaimerY := sigY + 600 + 110
	drawText(dc, "This ID card is valid for authorized person only,", float64(w)/2, disclaimerY, mustFont(fonts.AptosItalic, 100), hexColor("#000000"), 0.5, 0.5)
	drawText(dc, "If lost & found, please contact the above address", float64(w)/2, disclaimerY+120, mustFont(fonts.AptosItalic, 100), hexColor("#000000"), 0.5, 0.5)

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode back image: %w", err)
	}
	return out.Bytes(), nil
}
