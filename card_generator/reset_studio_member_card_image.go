package card_generator

import (
	"bytes"
	"employee-card-admin/assets/fonts"
	"employee-card-admin/dto"
	"employee-card-admin/models"
	"employee-card-admin/utils"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

// ---------------------------------------------------------------------------
// Reset Studio member card
//
// The whole layout is expressed in "design units": the Figma card is
// 253 x 202 units. Every coordinate below was measured from the shared design
// and is converted to pixels with ux / uy (horizontal / vertical) and uc
// (uniform, for fonts and borders). So the same code renders the full-size
// card (5060x4040) and the 650x518 thumbnail without any layout changes.
// ---------------------------------------------------------------------------

const (
	// Full-size card output (20 px per design unit)
	resetCardW = 3942 //6174
	resetCardH = 3942

	// Thumbnail output
	resetThumbnailW = 650
	resetThumbnailH = 518

	// Figma card size in design units
	resetDesignW = 253.0
	resetDesignH = 202.0

	// Colors sampled from the design
	resetCreamHex      = "#F3F3EB" // card background
	resetNameHex       = "#0A0A08" // member name
	resetOliveHex      = "#8B8D68" // terms text (matches flower green)
	resetOliveBoldHex  = "#7C7E58" // tel / address text
	resetFrameBrownHex = "#926E52" // photo + QR frame fallback border
	resetNfcHex        = "#000000" // NFC icon
)

const (
	resetFlowerBgFrontURL = "RESETS_STUDIO_UI/1791870390-reset_studio_Flower_Backgound-front.png"
	resetBgBackURL        = "RESETS_STUDIO_UI/1791870298-reset_studio_Backgound-back.png"
	resetLogoFrontURL     = "RESETS_STUDIO_UI/1791870363-reset_studio_LogoURL.png"
	resetLogoBackURL      = "RESETS_STUDIO_UI/1791870330-reset_studio_LogoURL_back.png"
	resetQRFrameURL       = "RESETS_STUDIO_UI/1791870451-reset_studio_QR-image-front.png"
	resetNfcURL           = "RESETS_STUDIO_UI/1791873383-reset_studio_NfcURL.png"
)

func isResetStudioCard(businessName, designation string) bool {
	if utils.IsYogiDesignation(designation) {
		return true
	}
	b := strings.ToLower(strings.TrimSpace(businessName))
	d := strings.ToLower(strings.TrimSpace(designation))
	return strings.Contains(b, "reset studio") || strings.Contains(b, "resetstudio") ||
		strings.Contains(d, "reset studio") || strings.Contains(d, "resetstudio")
}

// resetDrawContainLeft fits img inside the box keeping its aspect ratio,
// left-aligned and vertically centered (so a logo lockup starts exactly at x).
func resetDrawContainLeft(dc *gg.Context, img image.Image, x, y, w, h float64) {
	if img == nil || w <= 0 || h <= 0 {
		return
	}
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return
	}
	scale := math.Min(w/float64(b.Dx()), h/float64(b.Dy()))
	dw := int(math.Round(float64(b.Dx()) * scale))
	dh := int(math.Round(float64(b.Dy()) * scale))
	if dw <= 0 || dh <= 0 {
		return
	}
	dc.DrawImage(scaleImageHighQuality(img, dw, dh), int(math.Round(x)), int(math.Round(y+(h-float64(dh))/2.0)))
}

// resetContainRect returns the centered rect img occupies when drawn with
// drawImageContain into (x, y, w, h).
func resetContainRect(img image.Image, x, y, w, h float64) (float64, float64, float64, float64) {
	if img == nil {
		return x, y, w, h
	}
	b := img.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return x, y, w, h
	}
	scale := math.Min(w/float64(b.Dx()), h/float64(b.Dy()))
	dw := float64(b.Dx()) * scale
	dh := float64(b.Dy()) * scale
	return x + (w-dw)/2.0, y + (h-dh)/2.0, dw, dh
}

// resetFitName returns the name as one line (shrinking down to ~75% of base),
// or two lines if it still does not fit.
func resetFitName(dc *gg.Context, name string, maxWidth, baseSize, twoLineSize, minSize float64) ([]string, font.Face) {
	clean := strings.TrimSpace(name)
	if clean == "" {
		clean = "Aarusha Zahia"
	}

	step := math.Max(baseSize*0.04, 0.5)
	for size := baseSize; size >= baseSize*0.75; size -= step {
		face := mustFont(fonts.AptosRegular, size)
		dc.SetFontFace(face)
		if w, _ := dc.MeasureString(clean); w <= maxWidth {
			return []string{clean}, face
		}
	}

	words := strings.Fields(clean)
	if len(words) <= 1 {
		face := mustFont(fonts.AptosRegular, baseSize*0.75)
		return []string{clean}, face
	}

	for size := twoLineSize; size >= minSize; size -= step {
		face := mustFont(fonts.AptosRegular, size)
		dc.SetFontFace(face)

		splitIdx := 1
		for i := 1; i < len(words); i++ {
			if cw, _ := dc.MeasureString(strings.Join(words[:i], " ")); cw <= maxWidth {
				splitIdx = i
			} else {
				break
			}
		}
		line1 := strings.Join(words[:splitIdx], " ")
		line2 := strings.Join(words[splitIdx:], " ")
		w1, _ := dc.MeasureString(line1)
		w2, _ := dc.MeasureString(line2)
		if w1 <= maxWidth && w2 <= maxWidth {
			return []string{line1, line2}, face
		}
	}

	face := mustFont(fonts.AptosRegular, minSize)
	mid := len(words) / 2
	return []string{strings.Join(words[:mid], " "), strings.Join(words[mid:], " ")}, face
}

// renderResetStudioFront draws the front side at any size (full card or thumbnail).
func renderResetStudioFront(data employeeCardTemplateData, w, h int) image.Image {
	dc := gg.NewContext(w, h)

	ux := float64(w) / resetDesignW
	uy := float64(h) / resetDesignH
	uc := math.Min(ux, uy)

	// 1. Background: cream base + flower artwork (top-right and bottom)
	dc.SetColor(hexColor(resetCreamHex))
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()

	if flowerImg, err := loadImageFromURIOrURL(resetFlowerBgFrontURL); err == nil && flowerImg != nil {
		drawImageCover(dc, flowerImg, 0, 0, float64(w), float64(h))
	}

	// 2. Top-left: logo lockup "RESET | Reset Studio Member Card"
	logoX := 12.5 * ux
	logoY := 26.5 * uy
	logoW := 101.5 * ux
	logoH := 12.5 * uy

	if logoImg, err := loadImageFromURIOrURL(resetLogoFrontURL); err == nil && logoImg != nil {
		resetDrawContainLeft(dc, autoCropNonEmpty(logoImg), logoX, logoY, logoW, logoH)
	} else {
		drawText(dc, "RESET", logoX, logoY+logoH/2.0, mustFont(fonts.AptosRegular, 11.0*uc), hexColor(resetNameHex), 0, 0.5)
	}

	// 3. Top-right: NFC icon (black), right-aligned, sits on top of the flower
	nfcH := 22.5 * uc
	nfcW := 20.0 * uc
	nfcRight := 236.0 * ux
	nfcY := 22.5 * uy

	if nfcImg, err := loadImageFromURIOrURL(resetNfcURL); err == nil && nfcImg != nil {
		croppedNFC := autoCropNonEmpty(nfcImg)
		tintedNFC := tintImageToColor(croppedNFC, resetNfcHex)
		if croppedNFC.Bounds().Dy() > 0 {
			nfcW = nfcH * (float64(croppedNFC.Bounds().Dx()) / float64(croppedNFC.Bounds().Dy()))
		}
		scaledNFC := scaleImageHighQuality(tintedNFC, int(math.Round(nfcW)), int(math.Round(nfcH)))
		dc.DrawImage(scaledNFC, int(math.Round(nfcRight-nfcW)), int(math.Round(nfcY)))
	} else if nfcFallback, err := renderNFCIcon(resetNfcHex); err == nil && nfcFallback != nil {
		drawImageContain(dc, nfcFallback, nfcRight-nfcW, nfcY, nfcW, nfcH)
	}

	// 4. Right: frame + payment QR code (in place of the member photo)
	frameX := 146.5 * ux
	frameY := 77.5 * uy
	frameW := 93.5 * ux
	frameH := 97.5 * uy

	// Shrink to the frame image's actual drawn rect so the white fill never
	// spills outside the frame border.
	var frameImg image.Image
	if img, err := loadImageFromURIOrURL(resetQRFrameURL); err == nil && img != nil {
		frameImg = autoCropNonEmpty(img)
		frameX, frameY, frameW, frameH = resetContainRect(frameImg, frameX, frameY, frameW, frameH)
	}

	framePad := 1.7 * uc
	innerX := frameX + framePad
	innerY := frameY + framePad
	innerW := frameW - framePad*2.0
	innerH := frameH - framePad*2.0

	dc.SetColor(color.White)
	dc.DrawRectangle(innerX, innerY, innerW, innerH)
	dc.Fill()

	qrPad := 2.2 * uc
	if paymentQRImg, err := loadImageOrGeneratePaymentQRCode(data.PaymentQRCode, 1800); err == nil && paymentQRImg != nil {
		drawImageContain(dc, ensureBlackQRCodeImage(paymentQRImg), innerX+qrPad, innerY+qrPad, innerW-qrPad*2.0, innerH-qrPad*2.0)
	}

	if frameImg != nil {
		drawImageContain(dc, frameImg, frameX, frameY, frameW, frameH)
	} else {
		dc.SetColor(hexColor(resetFrameBrownHex))
		dc.SetLineWidth(1.6 * uc)
		dc.DrawRectangle(frameX+0.8*uc, frameY+0.8*uc, frameW-1.6*uc, frameH-1.6*uc)
		dc.Stroke()
	}

	// 5. Left-middle: member name
	textX := 12.5 * ux
	maxNameWidth := (146.5 - 12.5 - 8.0) * ux

	memberName := universityFallbackString(data.MemberName, "Aarusha Zahia")
	nameLines, nameFace := resetFitName(dc, memberName, maxNameWidth, 14.5*uc, 12.0*uc, 8.5*uc)

	if len(nameLines) == 1 {
		drawText(dc, nameLines[0], textX, 107.0*uy, nameFace, hexColor(resetNameHex), 0, 0.5)
	} else {
		lineGap := 14.0 * uy
		for i, line := range nameLines {
			drawText(dc, line, textX, 107.0*uy-lineGap/2.0+float64(i)*lineGap, nameFace, hexColor(resetNameHex), 0, 0.5)
		}
	}

	return dc.Image()
}

func drawResetStudioMemberCardFront(data employeeCardTemplateData) ([]byte, error) {
	img := renderResetStudioFront(data, resetCardW, resetCardH)

	var out bytes.Buffer
	if err := png.Encode(&out, img); err != nil {
		return nil, fmt.Errorf("encode reset studio front image: %w", err)
	}
	return out.Bytes(), nil
}

func drawResetStudioMemberCardBack(data employeeCardTemplateData) ([]byte, error) {
	w, h := resetCardW, resetCardH
	dc := gg.NewContext(w, h)

	ux := float64(w) / resetDesignW
	uy := float64(h) / resetDesignH
	uc := math.Min(ux, uy)

	// 1. Background
	dc.SetColor(hexColor(resetCreamHex))
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()

	if bgImg, err := loadImageFromURIOrURL(resetBgBackURL); err == nil && bgImg != nil {
		drawImageCover(dc, bgImg, 0, 0, float64(w), float64(h))
	}

	// 2. Top-left: logo lockup "RESET | Reset Studio Member Card"
	logoX := 9.5 * ux
	logoY := 31.5 * uy
	logoW := 154.5 * ux
	logoH := 17.5 * uy

	if logoImg, err := loadImageFromURIOrURL(resetLogoBackURL); err == nil && logoImg != nil {
		resetDrawContainLeft(dc, autoCropNonEmpty(logoImg), logoX, logoY, logoW, logoH)
	} else {
		drawText(dc, "RESET", logoX, logoY+logoH/2.0, mustFont(fonts.AptosRegular, 16.0*uc), hexColor(resetNameHex), 0, 0.5)
	}

	// 3. Right: QR frame + VDS / OneID QR code
	frameX := 143.0 * ux
	frameY := 63.5 * uy
	frameW := 94.0 * ux
	frameH := 98.5 * uy

	var qrFrameImg image.Image
	if img, err := loadImageFromURIOrURL(resetQRFrameURL); err == nil && img != nil {
		qrFrameImg = autoCropNonEmpty(img)
		frameX, frameY, frameW, frameH = resetContainRect(qrFrameImg, frameX, frameY, frameW, frameH)
	}

	frameBorder := 1.7 * uc
	innerX := frameX + frameBorder
	innerY := frameY + frameBorder
	innerW := frameW - frameBorder*2.0
	innerH := frameH - frameBorder*2.0

	dc.SetColor(color.White)
	dc.DrawRectangle(innerX, innerY, innerW, innerH)
	dc.Fill()

	qrPad := 2.2 * uc
	oneIDQRImg, err := loadImageOrGenerateVDSQRCode(data.OneIDQRCode, 2400)
	if err == nil && oneIDQRImg != nil {
		drawImageContain(dc, oneIDQRImg, innerX+qrPad, innerY+qrPad, innerW-qrPad*2.0, innerH-qrPad*2.0)
	}

	if qrFrameImg != nil {
		drawImageContain(dc, qrFrameImg, frameX, frameY, frameW, frameH)
	} else {
		dc.SetColor(hexColor(resetFrameBrownHex))
		dc.SetLineWidth(frameBorder)
		dc.DrawRectangle(frameX+frameBorder/2.0, frameY+frameBorder/2.0, frameW-frameBorder, frameH-frameBorder)
		dc.Stroke()
	}

	// 4. Left-middle: terms text (olive)
	textX := 8.25 * ux
	termsFont := mustFont(fonts.AptosRegular, 6.2*uc)
	termsLines := []string{
		"This card is the property of Reset Studio",
		"and must presented upon request.",
		"Use of this card is subject to the terms",
		"and conditions of membership.",
	}
	for i, line := range termsLines {
		drawText(dc, line, textX, (105.75+float64(i)*7.75)*uy, termsFont, hexColor(resetOliveHex), 0, 0.5)
	}

	// 5. Contact info (bold olive)
	phone := universityFallbackString(data.PhoneNumber, "012XXXXXXXX")
	addrLine1, addrLine2 := formatSaintsAddress(data.Address)

	contactLines := []string{
		"Tel. " + phone,
		"Address: " + addrLine1,
	}
	if addrLine2 != "" {
		contactLines = append(contactLines, addrLine2)
	}

	contactFont := mustFont(fonts.HelveticaBold, 6.2*uc)
	for i, line := range contactLines {
		drawText(dc, line, textX, (142.5+float64(i)*7.9)*uy, contactFont, hexColor(resetOliveBoldHex), 0, 0.5)
	}

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode reset studio back image: %w", err)
	}
	return out.Bytes(), nil
}

// BuildResetStudioMemberThumbnailImage generates the 650x518 thumbnail
// (same design as the front, rendered at thumbnail size with rounded corners).
func BuildResetStudioMemberThumbnailImage(data employeeCardTemplateData) ([]byte, error) {
	w := resetThumbnailW
	h := resetThumbnailH
	dc := gg.NewContext(w, h)

	front := renderResetStudioFront(data, w, h)

	dc.Push()
	dc.DrawRoundedRectangle(0, 0, float64(w), float64(h), 0.03*float64(w))
	dc.Clip()
	dc.DrawImage(front, 0, 0)
	dc.Pop()
	dc.ResetClip()

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode reset studio thumbnail image: %w", err)
	}
	return out.Bytes(), nil
}

// BuildResetStudioMemberCardImages builds the front (payment QR), back (OneID / VDS QR)
// and thumbnail images for Reset Studio members.
func BuildResetStudioMemberCardImages(
	employee models.Employee,
	membership models.Membership,
	merchant models.Merchant,
	dgepayCard dto.GenericResponse,
	generateVDS dto.GenerateVDSData,
) (UniversityCardImages, error) {
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
		MemberName:    universityFallbackString(employee.UserName, "Aarusha Zahia"),
		Designation:   universityFallbackString(membership.Designation, "Member"),
		RegNo:         universityFallbackString(membership.RefID, "XXXX XXXX 1234"),
		OneID:         universityFallbackString(employee.OneID, "N/A"),
		OneIDQRCode:   universityFallbackImage(generateVDS.ExtractedQRImage),
		PaymentQRCode: paymentQRValue,
		IssueDate:     branding.IssueDate,
		ExpiryDate:    branding.ExpiryDate,
		BusinessName:  universityFallbackString(merchant.BusinessName, "Reset Studio"),
		PhoneNumber:   universityFallbackString(merchant.Phone, "+880 1865-840108"),
		Address:       universityFallbackString(branding.Address, "Road 73, House 4D, 5th floor, Gulshan 2, Dhaka-1212"),
	}

	front, err := drawResetStudioMemberCardFront(data)
	if err != nil {
		return UniversityCardImages{}, err
	}
	back, err := drawResetStudioMemberCardBack(data)
	if err != nil {
		return UniversityCardImages{}, err
	}
	thumbnail, err := BuildResetStudioMemberThumbnailImage(data)
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
