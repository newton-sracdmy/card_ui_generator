package card_generator

import (
	"bytes"
	"employee-card-admin/assets/fonts"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

const (
	saintsThumbnailW = 650
	saintsThumbnailH = 518

	saintsThumbnailBgURL           = "SAINTS_CLUB_THUMBMAIL_UI/1791437868-background.png"
	saintsThumbnailLogoURL         = "SAINTS_CLUB_MEMBER_UI/1791207148-Logo@4x.png"
	saintsThumbnailNfcURL          = "SAINTS_CLUB_THUMBMAIL_UI/1791437956-NFC.png"
	saintsThumbnailPhotoFrameURL   = "SAINTS_CLUB_THUMBMAIL_UI/1791437977-Photo-Frame@4x-1.png"
	saintsThumbnailLowerLineURL    = "SAINTS_CLUB_THUMBMAIL_UI/1791438001-saint_lower-line.png"
	saintsThumbnailTitleURL        = "SAINTS_CLUB_THUMBMAIL_UI/1791443039-SAINT_CLUB_TITLE.png"
	saintsThumbnailBanglaQRLogoURL = "UAP_Student_Card_UI/1790672994-Bangla-QR.png"
)

func splitThumbnailMemberName(dc *gg.Context, name string, maxWidth float64) ([]string, font.Face) {
	clean := strings.TrimSpace(name)
	if clean == "" {
		face := mustFont(fonts.AptosRegular, 30.0)
		return []string{"Aarusha Zahia"}, face
	}

	curSize := 32.0
	minSize := 20.0
	for curSize >= minSize {
		face := mustFont(fonts.AptosRegular, curSize)
		dc.SetFontFace(face)
		w, _ := dc.MeasureString(clean)
		if w <= maxWidth {
			return []string{clean}, face
		}

		words := strings.Fields(clean)
		if len(words) <= 1 {
			curSize -= 2.0
			continue
		}

		splitIdx := 1
		for i := 1; i < len(words); i++ {
			cand := strings.Join(words[:i], " ")
			cw, _ := dc.MeasureString(cand)
			if cw <= maxWidth {
				splitIdx = i
			} else {
				break
			}
		}

		line1 := strings.Join(words[:splitIdx], " ")
		line2 := strings.Join(words[splitIdx:], " ")

		w2, _ := dc.MeasureString(line2)
		if w2 <= maxWidth {
			return []string{line1, line2}, face
		}

		curSize -= 2.0
	}

	face := mustFont(fonts.AptosRegular, minSize)
	dc.SetFontFace(face)
	words := strings.Fields(clean)
	if len(words) <= 1 {
		return []string{clean}, face
	}
	splitIdx := len(words) / 2
	return []string{strings.Join(words[:splitIdx], " "), strings.Join(words[splitIdx:], " ")}, face
}

// BuildSaintsClubMemberThumbnailImage generates the 650x518 thumbnail card image
func BuildSaintsClubMemberThumbnailImage(data employeeCardTemplateData) ([]byte, error) {
	w := saintsThumbnailW
	h := saintsThumbnailH
	dc := gg.NewContext(w, h)

	cardW := float64(w)
	cardH := float64(h)
	cardR := 32.0

	// 1. Draw rounded background & lower bar clipped to card rounded rectangle
	dc.Push()
	dc.DrawRoundedRectangle(0, 0, cardW, cardH, cardR)
	dc.Clip()

	if bgImg, err := loadImageFromURIOrURL(saintsThumbnailBgURL); err == nil && bgImg != nil {
		dc.DrawImage(scaleImageHighQuality(bgImg, int(cardW), int(cardH)), 0, 0)
	} else {
		dc.SetColor(hexColor("#0E0E12"))
		dc.DrawRectangle(0, 0, cardW, cardH)
		dc.Fill()
	}

	barH := 16.0
	barY := cardH - barH
	if barImg, err := loadImageFromURIOrURL(saintsThumbnailLowerLineURL); err == nil && barImg != nil {
		dc.DrawImage(scaleImageHighQuality(barImg, int(cardW), int(barH)), 0, int(barY))
	} else {
		dc.SetColor(hexColor("#BC925C"))
		dc.DrawRectangle(0, barY, cardW, barH)
		dc.Fill()
	}
	dc.Pop()
	dc.ResetClip()

	// 2. Top-Left: Logo & Main Titles
	logoX := 26.0
	logoY := 52.0
	logoSize := 71.0

	if logoImg, err := loadImageFromURIOrURL(saintsThumbnailLogoURL); err == nil && logoImg != nil {
		croppedLogo := autoCropNonEmpty(logoImg)
		drawImageContain(dc, croppedLogo, logoX, logoY, logoSize, logoSize)
	}

	titleX := 102.0
	titleY := 68.0
	titleW := 285.0
	titleH := 41.0

	if titleImg, err := loadImageFromURIOrURL(saintsThumbnailTitleURL); err == nil && titleImg != nil {
		croppedTitle := autoCropNonEmpty(titleImg)
		drawImageContain(dc, croppedTitle, titleX, titleY, titleW, titleH)
	} else {
		drawVerticalGradientText(dc, "SAINTS CLUB MEMBER CARD", titleX, titleY+titleH*0.35, mustFont(fonts.HelveticaBold, 15.0), "#C99032", "#EAE2B1", 0.202, 0, 0.5)
		drawVerticalGradientText(dc, "ONE CLUB ALLIANCE", titleX, titleY+titleH*0.75, mustFont(fonts.HelveticaBold, 12.0), "#C99032", "#EAE2B1", 0.202, 0, 0.5)
	}

	// 3. Top-Right: NFC Icon (tinted white)
	nfcX := 535.0
	nfcY := 56.0
	nfcW := 46.0
	nfcH := 62.0

	if nfcImg, err := loadImageFromURIOrURL(saintsThumbnailNfcURL); err == nil && nfcImg != nil {
		croppedNFC := autoCropNonEmpty(nfcImg)
		tintedNFC := tintImageToColor(croppedNFC, "#FFFFFF")
		scaledNFC := scaleImageHighQuality(tintedNFC, int(math.Round(nfcW)), int(math.Round(nfcH)))
		dc.DrawImage(scaledNFC, int(math.Round(nfcX)), int(math.Round(nfcY)))
	} else if nfcFallback, err := renderNFCIcon("#FFFFFF"); err == nil && nfcFallback != nil {
		drawImageContain(dc, nfcFallback, nfcX, nfcY, nfcW, nfcH)
	}

	// 4. Left-Middle: Photo Frame & Framed Avatar
	frameX := 34.0
	frameY := 144.0
	frameW := 92.0
	frameH := 107.0

	photoPad := 4.0
	photoX := frameX + photoPad
	photoY := frameY + photoPad
	photoW := frameW - photoPad*2.0
	photoH := frameH - photoPad*2.0

	drawRectangularAvatar(dc, photoX, photoY, photoW, photoH, data.AvatarURL, data.Initials)

	if frameImg, err := loadImageFromURIOrURL(saintsThumbnailPhotoFrameURL); err == nil && frameImg != nil {
		croppedFrame := autoCropNonEmpty(frameImg)
		drawImageContain(dc, croppedFrame, frameX, frameY, frameW, frameH)
	} else {
		dc.SetColor(hexColor("#D4AF37"))
		dc.SetLineWidth(3.0)
		dc.DrawRectangle(photoX, photoY, photoW, photoH)
		dc.Stroke()
	}

	// 5. Right-Middle: Payment QR Code
	qrPanelSize := 192.0
	qrPanelX := 402.0
	qrPanelY := 280.0
	qrPanelRadius := 12.0

	drawRoundedPanel(dc, qrPanelX, qrPanelY, qrPanelSize, qrPanelSize, qrPanelRadius, color.White)

	qrPadding := 8.0
	qrContentX := qrPanelX + qrPadding
	qrContentY := qrPanelY + qrPadding
	qrContentSize := qrPanelSize - qrPadding*2.0

	qrImg, err := loadImageOrGeneratePaymentQRCode(data.PaymentQRCode, 800)
	if err == nil && qrImg != nil {
		blackQR := ensureBlackQRCodeImage(qrImg)
		drawImageContain(dc, blackQR, qrContentX, qrContentY, qrContentSize, qrContentSize)
	}

	// Bangla QR Center Logo
	banglaSize := 38.0
	banglaCenterX := qrPanelX + qrPanelSize/2.0
	banglaCenterY := qrPanelY + qrPanelSize/2.0
	banglaRadius := 4.0

	drawRoundedPanel(dc, banglaCenterX-banglaSize/2.0, banglaCenterY-banglaSize/2.0, banglaSize, banglaSize, banglaRadius, color.White)

	if banglaImg, err := loadImageFromURIOrURL(saintsThumbnailBanglaQRLogoURL); err == nil && banglaImg != nil {
		drawImageContain(dc, banglaImg, banglaCenterX-banglaSize/2.0+2.0, banglaCenterY-banglaSize/2.0+2.0, banglaSize-4.0, banglaSize-4.0)
	} else {
		drawText(dc, "BANGLA", banglaCenterX, banglaCenterY-3.0, mustFont(fonts.HelveticaBold, 5.0), hexColor("#E31D2B"), 0.5, 0.5)
		drawText(dc, "QR", banglaCenterX, banglaCenterY+4.0, mustFont(fonts.HelveticaBold, 6.0), hexColor("#00A859"), 0.5, 0.5)
	}

	// 6. Left Column: Member Name & Membership ID
	detailsStartX := 35.0
	memberName := universityFallbackString(data.MemberName, "Aarusha Zahia")
	maxNameWidth := qrPanelX - detailsStartX - 20.0

	nameLines, nameFace := splitThumbnailMemberName(dc, memberName, maxNameWidth)

	var labelY, valY float64
	labelFont := mustFont(fonts.AptosRegular, 17.0)
	valueFont := mustFont(fonts.AptosRegular, 22.0)

	if len(nameLines) == 1 {
		nameY := 295.0
		drawVerticalGradientText(dc, nameLines[0], detailsStartX, nameY, nameFace, "#C99032", "#EAE2B1", 0.202, 0, 0.5)
		labelY = 349.0
		valY = 386.0
	} else {
		nameLineSpacing := 30.0
		nameStartY := 280.0
		drawVerticalGradientLines(dc, nameLines, detailsStartX, nameStartY, nameLineSpacing, nameFace, "#C99032", "#EAE2B1", 0.202, 0, 0.5)
		labelY = 354.0
		valY = 389.0
	}

	drawVerticalGradientText(dc, "Membership ID:", detailsStartX+2.0, labelY, labelFont, "#C99032", "#EAE2B1", 0.202, 0, 0.5)
	memIDStr := formatCardNumberOrRef(universityFallbackString(data.RegNo, "XXXX XXXX 1234"))
	drawVerticalGradientText(dc, memIDStr, detailsStartX+2.0, valY, valueFont, "#C99032", "#EAE2B1", 0.202, 0, 0.5)

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode saints club thumbnail image: %w", err)
	}
	return out.Bytes(), nil
}
