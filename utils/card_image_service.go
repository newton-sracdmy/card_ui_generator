package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"time"

	"employee-card-admin/dto"
	"employee-card-admin/models"

	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/fogleman/gg"
	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
)

const (
	cardFrontW   = 416
	cardFrontH   = 660
	cardBackW    = 414
	cardBackH    = 656
	cardWidthPx  = 416
	cardHeightPx = 660

	bloodIconB64 = "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"
	nfcIconB64   = "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"
)

type CardImages struct {
	// OneID and Payment are image/svg+xml bytes (PNG raster wrapped in SVG for delivery).
	OneID   []byte
	Payment []byte
	// oneIDPNGForSync is the raw PNG raster sent to the OneID document sync API.
	oneIDPNGForSync []byte
}

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

type employeeCardTemplateData struct {
	MemberName          string
	Designation         string
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
	Fax                 string
	WebsiteURL          string
	Address             string
	BrandColor          string
	SignatureURL        string
	SignatureLabel      string
	FooterDisclaimer1   string
	FooterDisclaimer2   string
}

type cardBranding struct {
	BrandColor          string
	Address             string
	OptionalPhoneNumber string
	Fax                 string
	WebsiteURL          string
	SignatureURL        string
	IssueDate           string
	ExpiryDate          string
}

const defaultBrandColor = "#631FB4"

func  buildEmployeeCardImages(
	employee models.Employee,
	membership models.Membership,
	merchant models.Merchant,
	dgepayCard dto.GenericResponse,
	generateVDS dto.GenerateVDSData,
) (CardImages, error) {
	paymentData, err := extractDgepayCardData(dgepayCard)
	if err != nil {
		return CardImages{}, err
	}

	paymentQRStr := strings.TrimSpace(paymentData.QRCode)
	var paymentQRValue string
	if strings.HasPrefix(paymentQRStr, "http://") || strings.HasPrefix(paymentQRStr, "https://") || strings.HasPrefix(paymentQRStr, "data:") {
		paymentQRValue = paymentQRStr
	} else {
		paymentQR, err := qrcode.Encode(paymentQRStr, qrcode.Medium, 512)
		if err != nil {
			return CardImages{}, fmt.Errorf("generate payment qr: %w", err)
		}
		paymentQRValue = pngDataURI(paymentQR)
	}

	data := buildEmployeeCardTemplateData(employee, membership, merchant, paymentData, generateVDS, paymentQRValue)

	oneIDSVG, oneIDPNG, err := drawOneIDCard(data)
	if err != nil {
		return CardImages{}, err
	}

	paymentSVG, err := drawPaymentCard(data)
	if err != nil {
		return CardImages{}, err
	}

	return CardImages{
		OneID:           oneIDSVG,
		Payment:         paymentSVG,
		oneIDPNGForSync: oneIDPNG,
	}, nil
}

func buildEmployeeCardTemplateData(
	employee models.Employee,
	membership models.Membership,
	merchant models.Merchant,
	paymentData dgepayCardData,
	generateVDS dto.GenerateVDSData,
	paymentQRValue string,
) employeeCardTemplateData {
	branding := resolveCardBranding(merchant, paymentData.IssuedAt, paymentData.ExpiryDate)

	phoneNumber := fallbackString(paymentData.UserInfo.PhoneNumber, employee.Phone)
	if phoneNumber == "" || phoneNumber == "N/A" {
		phoneNumber = fallbackString(merchant.Phone, "N/A")
	}

	return employeeCardTemplateData{
		MemberName:  fallbackString(employee.UserName, "N/A"),
		Designation: fallbackString(membership.Designation, "Employee"),
		OneID:       fallbackString(employee.OneID, "N/A"),
		Initials:    buildInitials(employee.UserName),
		BloodGroup:  fallbackString(GetStringFromMetaData(membership.MetaData, "blood_group"), "A+"),

		OneIDQRCode:   fallbackImage(generateVDS.ExtractedQRImage),
		AvatarURL:     employee.AvatarURL,
		PaymentQRCode: paymentQRValue,
		IssueDate:     branding.IssueDate,
		ExpiryDate:    branding.ExpiryDate,
		BusinessName: strings.ToUpper(
			fallbackString(paymentData.MerchantInfo.BusinessName, merchant.BusinessName),
		),
		PhoneNumber:         phoneNumber,
		OptionalPhoneNumber: branding.OptionalPhoneNumber,
		Fax:                 branding.Fax,
		WebsiteURL:          branding.WebsiteURL,
		Address:             branding.Address,
		BrandColor:          branding.BrandColor,
		SignatureURL:        branding.SignatureURL,
		SignatureLabel:      "Authorised Signature",
		FooterDisclaimer1:   "This ID card is valid for authorized person only.",
		FooterDisclaimer2:   "If lost & found, please contact the above address.",
	}
}

func resolveCardBranding(merchant models.Merchant, cardIssued, cardExpiry any) cardBranding {
	branding := cardBranding{
		BrandColor: defaultBrandColor,
		Address:    "N/A",
		IssueDate:  formatCardDate(cardIssued),
		ExpiryDate: formatCardDate(cardExpiry),
	}

	if merchant.Settings != nil {
		settings := merchant.Settings
		if color := strings.TrimSpace(settings.BrandColor); color != "" {
			branding.BrandColor = color
		}
		if addr := strings.TrimSpace(settings.Address); addr != "" {
			branding.Address = addr
		}
		branding.SignatureURL = strings.TrimSpace(settings.Signature)
		branding.OptionalPhoneNumber = optionalStringValue(settings.OptionalPhoneNumber)
		branding.Fax = optionalStringValue(settings.Fax)
		branding.WebsiteURL = optionalStringValue(settings.WebsiteURL)
		if !settings.CreatedAt.IsZero() {
			branding.IssueDate = settings.CreatedAt.Format("02 January 2006")
		}
		if !settings.CardExpiryTime.IsZero() {
			branding.ExpiryDate = settings.CardExpiryTime.Format("02 January 2006")
		}
	}

	return branding
}

func optionalStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func drawOneIDCard(data employeeCardTemplateData) ([]byte, []byte, error) {
	const w, h = cardFrontW, cardFrontH
	dc := gg.NewContext(w, h)
	brandColor := hexColor(data.BrandColor)
	drawCardBackground(dc, brandColor)

	// ── White header panel (logo area) ──
	// CSS: left=10.9%, right=10.67%, top=5.82%, bottom=77.85%
	panelX := 0.109 * float64(w)  // 45
	panelY := 0.0582 * float64(h) // 38
	panelW := 0.7843 * float64(w) // 326
	panelH := 0.1633 * float64(h) // 108
	drawRoundedPanel(dc, panelX, panelY, panelW, panelH, 8, color.White)

	// Initials badge or avatar overriding bottom of header
	ddrawAvatarImage(dc, panelX+72, panelY+panelH-40, 62, data.AvatarURL, data.Initials, data.BrandColor)

	drawText(dc, data.BusinessName, 0.49*float64(w), panelY+panelH/2, mustFont(gobold.TTF, 11), brandColor, 0.5, 0.5)

	// ── Member name & designation ──
	// CSS: left=10.61%, top=27.95% and top=31.24%, font-size=9px→18pt
	drawText(dc, strings.ToUpper(data.MemberName), 0.1061*float64(w), 0.2795*float64(h),
		mustFont(gobold.TTF, 18), color.White, 0, 0.5)
	drawText(dc, data.Designation, 0.1061*float64(w), 0.3124*float64(h),
		mustFont(gobold.TTF, 18), color.White, 0, 0.5)

	// ── White info panel (Personal Information + QR) ──
	// CSS: left=10.52%, right=10.49%, top=38.8%, bottom=6.12%
	infoX := 0.1052 * float64(w) // 44
	infoY := 0.388 * float64(h)  // 256
	infoW := 0.7899 * float64(w) // 329
	infoH := 0.5508 * float64(h) // 364
	drawRoundedPanel(dc, infoX, infoY, infoW, infoH, 10, color.White)

	// "Personal Information" header
	// CSS: left=29.31%, top=40.95%, font-size=6.33px→12.66pt, color=#5E1EB2
	drawText(dc, "Personal Information", 0.5*float64(w), 0.4095*float64(h)+9,
		mustFont(gobold.TTF, 12.66), brandColor, 0.5, 0.5)

	// QR code
	// CSS clip: left=14.37%, right=14.8%, top=45.71%, bottom=9.75%
	qrX := 0.1437 * float64(w) // 60
	qrY := 0.4571 * float64(h) // 302
	qrW := 0.7083 * float64(w) // 295
	qrH := 0.445 * float64(h)  // 294
	qrImg, err := decodeDataURIToImage(data.OneIDQRCode)
	if err == nil {
		drawImageContain(dc, qrImg, qrX, qrY, qrW, qrH)
	}

	// ── Blood group badge (use blood.png icon) ──
	bgX := 0.80 * float64(w)
	bgY := 0.42 * float64(h)
	bloodImg := loadEmbeddedPNG(bloodIconB64)
	if bloodImg != nil {
		drawImageContain(dc, bloodImg, bgX-22, bgY-22, 44, 44)
	}
	drawText(dc, data.BloodGroup, bgX, bgY, mustFont(gobold.TTF, 14), color.White, 0.5, 0.5)

	// ── Vertical ID text (rotated -90°) ──
	// CSS: left=85.88%, top=32.75%, transform: rotate(-90deg)
	// Place just to the right of the info panel with some margin
	dc.Push()
	dc.Translate(infoX+infoW+14, 0.65*float64(h))
	dc.Rotate(-math.Pi / 2)
	drawText(dc, "ID - "+data.OneID, 0, 0, mustFont(goregular.TTF, 20), color.White, 0.5, 0.5)
	dc.Pop()
	dc.ResetClip()

	return encodeRasterCardAsSVG(dc, cardFrontW, cardFrontH)
}

// encodeRasterCardAsSVG renders the canvas to PNG once, then returns SVG wrapping that PNG (for .svg delivery) and the raw PNG (e.g. for OneID sync).
func encodeRasterCardAsSVG(dc *gg.Context, width, height int) (svg []byte, pngOut []byte, err error) {
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, dc.Image()); err != nil {
		return nil, nil, fmt.Errorf("raster encode: %w", err)
	}
	pngOut = append([]byte(nil), pngBuf.Bytes()...)
	b64 := base64.StdEncoding.EncodeToString(pngOut)
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>`)
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height)
	fmt.Fprintf(&b, `<image width="%d" height="%d" href="data:image/png;base64,%s"/>`, width, height, b64)
	b.WriteString(`</svg>`)
	return []byte(b.String()), pngOut, nil
}

func drawPaymentCard(data employeeCardTemplateData) ([]byte, error) {
	const w, h = cardBackW, cardBackH
	dc := gg.NewContext(w, h)
	brandColor := hexColor(data.BrandColor)
	accentColor := brandAccentColor(data.BrandColor)
	drawCardBackground(dc, brandColor)

	// ── White QR panel ──
	// CSS: left=11.05%, right=9.39%, top=9.09%, bottom=57.96%
	pnlX := 0.1105 * float64(w) // 46
	pnlY := 0.0909 * float64(h) // 60
	pnlW := 0.7956 * float64(w) // 329
	pnlH := 0.3295 * float64(h) // 216
	drawRoundedPanel(dc, pnlX, pnlY, pnlW, pnlH, 8, color.White)

	// QR code inside panel
	// CSS clip: left=13.8%, right=39.48%, top=11.07%, bottom=59.56%
	qrImg, err := decodeDataURIToImage(data.PaymentQRCode)
	if err != nil {
		return nil, fmt.Errorf("decode payment qr: %w", err)
	}
	qrX := 0.138 * float64(w)  // 57
	qrY := 0.1107 * float64(h) // 73
	qrW := 0.4672 * float64(w) // 193
	qrH := 0.2937 * float64(h) // 193
	drawImageContain(dc, qrImg, qrX, qrY, qrW, qrH)

	// "Payment" text
	// CSS: left=65.75%, top=9.41%, font-size=12px→24pt, bold, brand color
	drawText(dc, "Payment", 0.6575*float64(w), 0.0941*float64(h)+12,
		mustFont(gobold.TTF, 24), brandColor, 0, 0.5)

	// NFC icon (use nfc.png)
	// CSS: center ~(316, 174), radius ~35, fill=#631FB4
	nfcX := 0.7627 * float64(w) // 316
	nfcY := 0.2648 * float64(h) // 174
	nfcImg := loadEmbeddedPNG(nfcIconB64)
	if nfcImg != nil {
		drawImageContain(dc, nfcImg, nfcX-35, nfcY-35, 70, 70)
	}

	// ── Issue / Expiry dates ──
	// CSS: Issue Date left=10.35%, top=43.08%, font=6.3px→12.6pt, #C7A4FF
	// CSS: Expiry Date left=10.35%, top=46.1%
	// CSS: colon ":" left=35.07%
	// CSS: values left=41.1%
	labelX := 0.1035 * float64(w) // 43
	colonX := 0.3507 * float64(w) // 145
	valueX := 0.411 * float64(w)  // 170
	issueY := 0.4308 * float64(h) // 283
	expiryY := 0.461 * float64(h) // 302

	drawText(dc, "Issue Date", labelX, issueY, mustFont(goregular.TTF, 12.6), accentColor, 0, 0.5)
	drawText(dc, ":", colonX, issueY, mustFont(goregular.TTF, 12.6), color.White, 0, 0.5)
	drawText(dc, data.IssueDate, valueX, issueY, mustFont(goregular.TTF, 12.6), color.White, 0, 0.5)

	drawText(dc, "Expiry Date", labelX, expiryY, mustFont(goregular.TTF, 12.6), accentColor, 0, 0.5)
	drawText(dc, ":", colonX, expiryY, mustFont(goregular.TTF, 12.6), color.White, 0, 0.5)
	drawText(dc, data.ExpiryDate, valueX, expiryY, mustFont(goregular.TTF, 12.6), color.White, 0, 0.5)

	// ── Company info (optional contact lines only when values exist) ──
	contactX := 0.1678 * float64(w)
	contactY := 0.5548 * float64(h)
	lineStep := 0.032 * float64(h)

	drawText(dc, data.BusinessName, 0.1149*float64(w), contactY,
		mustFont(gobold.TTF, 15.1), accentColor, 0, 0.5)
	contactY += lineStep

	if phone := strings.TrimSpace(data.PhoneNumber); phone != "" && phone != "N/A" {
		drawText(dc, phone, contactX, contactY, mustFont(gobold.TTF, 10.8), color.White, 0, 0.5)
		contactY += lineStep
	}
	if optionalPhone := strings.TrimSpace(data.OptionalPhoneNumber); optionalPhone != "" {
		drawText(dc, optionalPhone, contactX, contactY, mustFont(gobold.TTF, 10.8), color.White, 0, 0.5)
		contactY += lineStep
	}
	if fax := strings.TrimSpace(data.Fax); fax != "" {
		drawText(dc, "Fax: "+fax, contactX, contactY, mustFont(gobold.TTF, 10.8), color.White, 0, 0.5)
		contactY += lineStep
	}
	if website := strings.TrimSpace(data.WebsiteURL); website != "" {
		drawText(dc, website, contactX, contactY, mustFont(gobold.TTF, 10.8), color.White, 0, 0.5)
		contactY += lineStep
	}
	if address := strings.TrimSpace(data.Address); address != "" && address != "N/A" {
		dc.SetFontFace(mustFont(goregular.TTF, 8.9))
		dc.SetColor(color.White)
		dc.DrawStringWrapped(address, 0.167*float64(w), contactY, 0, 0, 260, 1.4, gg.AlignLeft)
	}

	// ── Signature area (white strip) ──
	// CSS: top=77.14%, bottom=13.41% → y=506, h=62
	sigY := 0.7600 * float64(h) // 506
	sigH := 0.0945 * float64(h) // 62
	drawRoundedPanel(dc, 0, sigY, float64(w), sigH, 0, color.White)

	// Signature image or line
	lineY := sigY + sigH*0.65
	if !drawSignatureImage(dc, data.SignatureURL, 20, sigY+8, 120, sigH-16) {
		dc.SetColor(hexColor("#444444"))
		dc.SetLineWidth(1)
		dc.DrawLine(20, lineY, 140, lineY)
		dc.Stroke()
	}

	// "Authorised Signature" text
	// CSS: left=54.09%, top=80.7%, font=4.65px→9.3pt, #0B0B0B
	drawText(dc, data.SignatureLabel, 0.5409*float64(w), 0.807*float64(h),
		mustFont(gobold.TTF, 9.3), hexColor("#0B0B0B"), 0, 0.5)

	// ── Footer disclaimers (white, bold italic, large) ──
	// CSS: font-size=12px at card scale → 12pt on 2x canvas
	// CSS: top=85.43% and top=87.9%
	dc.SetFontFace(mustFont(gobold.TTF, 12))
	dc.SetColor(color.White)
	dc.DrawStringWrapped(
		data.FooterDisclaimer1+"\n"+data.FooterDisclaimer2,
		float64(w)/2, 0.8543*float64(h), 0.5, 0, float64(w)*0.75, 1.5, gg.AlignCenter,
	)

	svg, _, encErr := encodeRasterCardAsSVG(dc, cardBackW, cardBackH)
	if encErr != nil {
		return nil, fmt.Errorf("encode payment image: %w", encErr)
	}
	return svg, nil
}

func drawCardBackground(dc *gg.Context, brand color.Color) {
	w := dc.Width()
	h := dc.Height()
	dc.SetColor(brand)
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()

	// Dot strips at ~0.3% → ~5% top and ~95% → ~99.3% bottom (from CSS)
	drawDotStrip(dc, w, 0, 28)
	drawDotStrip(dc, w, h-28, 28)
}

func drawDotStrip(dc *gg.Context, cardW, startY, stripH int) {
	// CSS: opacity 0.25 white dots, spacing ~1.23% of card width
	dc.SetColor(withAlpha(hexColor("#FFFFFF"), 0.25))
	spacing := int(float64(cardW) * 0.0123) // ~5px
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

func drawInitialsBadge(dc *gg.Context, x, y, radius float64, initials, brandColor string) {
	drawCircle(dc, x, y, radius, brandAccentColor(brandColor))
	fontSize := radius * 0.75
	drawText(dc, fallbackString(initials, "NA"), x, y, mustFont(gobold.TTF, fontSize), hexColor(brandColor), 0.5, 0.5)
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

	dc.Push()
	dc.Translate(posX, posY)
	dc.Scale(scale, scale)
	dc.DrawImage(img, 0, 0)
	dc.Pop()
	dc.ResetClip()
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

func extractDgepayCardData(resp dto.GenericResponse) (dgepayCardData, error) {
	var envelope dgepayCardEnvelope

	raw, err := json.Marshal(resp.Data)
	if err != nil {
		return dgepayCardData{}, fmt.Errorf("marshal dgepay response: %w", err)
	}

	if err := json.Unmarshal(raw, &envelope); err != nil {
		return dgepayCardData{}, fmt.Errorf("parse dgepay response: %w", err)
	}

	if strings.TrimSpace(envelope.UserEmpCardsData.QRCode) == "" {
		return dgepayCardData{}, fmt.Errorf("payment qr code is missing")
	}

	return envelope.UserEmpCardsData, nil
}

func pngDataURI(pngBytes []byte) string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(pngBytes)
}

func fallbackImage(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "data:image/svg+xml;base64,PHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHdpZHRoPSIyNDAiIGhlaWdodD0iMjQwIj48cmVjdCB3aWR0aD0iMTAwJSUiIGhlaWdodD0iMTAwJSUiIGZpbGw9IiNmNWY1ZjUiLz48dGV4dCB4PSI1MCUiIHk9IjUwJSUiIGRvbWluYW50LWJhc2VsaW5lPSJtaWRkbGUiIHRleHQtYW5jaG9yPSJtaWRkbGUiIGZpbGw9IiM4ODgiIGZvbnQtc2l6ZT0iMTYiPk5vIFFSPC90ZXh0Pjwvc3ZnPg=="
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

func brandAccentColor(brandHex string) color.Color {
	base := hexColor(brandHex)
	r8, g8, b8, _ := base.RGBA()
	const whiteBlend = 0.55
	blendChannel := func(channel uint32) uint8 {
		channelValue := float64(channel >> 8)
		return uint8(channelValue*(1-whiteBlend) + 255*whiteBlend)
	}
	return color.RGBA{
		R: blendChannel(r8),
		G: blendChannel(g8),
		B: blendChannel(b8),
		A: 255,
	}
}

func drawSignatureImage(dc *gg.Context, imageURL string, x, y, w, h float64) bool {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return false
	}

	img, err := loadImageFromURL(imageURL)
	if err != nil {
		return false
	}

	drawImageContain(dc, img, x, y, w, h)
	return true
}

func loadImageFromURL(imageURL string) (image.Image, error) {
	if strings.HasPrefix(imageURL, "data:") {
		return decodeDataURIToImage(imageURL)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(imageURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	return img, err
}

func mustFont(ttf []byte, size float64) font.Face {
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
	if len(value) == 8 {
		value = value[:6]
	}
	if len(value) != 6 {
		return hexColor(defaultBrandColor)
	}

	var r, g, b uint8
	if _, err := fmt.Sscanf(value, "%02x%02x%02x", &r, &g, &b); err != nil {
		return hexColor(defaultBrandColor)
	}
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
func ddrawAvatarImage(dc *gg.Context, x, y, radius float64, badgeURL, initials, brandColor string) {
	badgeURL = strings.TrimSpace(badgeURL)
	if badgeURL != "" {
		img, err := loadImageFromURL(badgeURL)
		if err == nil {
			dc.Push()
			dc.DrawCircle(x, y, radius)
			dc.Clip()
			bounds := img.Bounds()
			imgW := float64(bounds.Dx())
			imgH := float64(bounds.Dy())

			scale := radius * 2 / imgW
			if imgH*scale < radius*2 {
				scale = radius * 2 / imgH
			}

			dc.Translate(x-imgW*scale/2, y-imgH*scale/2)
			dc.Scale(scale, scale)
			dc.DrawImage(img, 0, 0)
			dc.Pop()
			dc.ResetClip()

			dc.DrawCircle(x, y, radius)
			dc.SetColor(brandAccentColor(brandColor))
			dc.SetLineWidth(4)
			dc.Stroke()
			return
		}
	}
	drawInitialsBadge(dc, x, y, radius, initials, brandColor)
}
