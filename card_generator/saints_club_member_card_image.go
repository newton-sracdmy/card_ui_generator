package card_generator

import (
	"bytes"
	"employee-card-admin/assets/fonts"
	"employee-card-admin/dto"
	"employee-card-admin/models"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"
	"time"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
)

const (
	saintsCardW      = 6337
	saintsCardH      = 4101
	saintsBgURL      = "SAINTS_CLUB_MEMBER_UI/1791206886-Backgound@4x.png"
	saintsFrontBgURL = "SAINTS_CLUB_MEMBER_UI/1791206886-Backgound@4x.png"
	//saintsBackBgURL       = "SAINTS_CLUB_MEMBER_UI/1791219492-Backgound-copy@4x.png"
	saintsBackBgURL       = "SAINTS_CLUB_MEMBER_UI/1791206886-Backgound@4x.png"
	saintsLowerBarURL     = "SAINTS_CLUB_MEMBER_UI/1791207203-Lower-bar@4x.png"
	saintsLogoURL         = "SAINTS_CLUB_MEMBER_UI/1791207148-Logo@4x.png"
	saintsMainTitlesURL   = "SAINTS_CLUB_MEMBER_UI/1791207242-Main-Titles@4x.png"
	saintsLineURL         = "SAINTS_CLUB_MEMBER_UI/1791206996-Line-below-the-text@4x.png"
	saintsPhotoFrameURL   = "SAINTS_CLUB_MEMBER_UI/1791207325-Photo-Frame@4x.png"
	saintsVDSFrameURL     = "SAINTS_CLUB_MEMBER_UI/1791207361-VDS-Frame@4x.png"
	saintsBanglaQRLogoURL = "UAP_Student_Card_UI/1790672994-Bangla-QR.png"
	saintsNfcURL          = "UAP_Student_Card_UI/1790672287-NFC.png"
)

func isSaintsClubCard(businessName, designation string) bool {
	b := strings.ToLower(strings.TrimSpace(businessName))
	d := strings.ToLower(strings.TrimSpace(designation))
	return strings.Contains(b, "saint") || strings.Contains(b, "saints club") ||
		strings.Contains(d, "saint") || strings.Contains(d, "saints club")
}

func cropImageRect(src image.Image, rect image.Rectangle) image.Image {
	type subImager interface {
		SubImage(r image.Rectangle) image.Image
	}
	if si, ok := src.(subImager); ok {
		return si.SubImage(rect)
	}
	dst := image.NewRGBA(image.Rect(0, 0, rect.Dx(), rect.Dy()))
	draw.Draw(dst, dst.Bounds(), src, rect.Min, draw.Src)
	return dst
}

func autoCropNonEmpty(src image.Image) image.Image {
	if src == nil {
		return nil
	}
	bounds := src.Bounds()
	minX, minY, maxX, maxY := bounds.Max.X, bounds.Max.Y, bounds.Min.X, bounds.Min.Y
	found := false
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, a := src.At(x, y).RGBA()
			if a > 2000 {
				found = true
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}
	if !found {
		return src
	}
	return cropImageRect(src, image.Rect(minX, minY, maxX+1, maxY+1))
}

func drawImageCover(dc *gg.Context, img image.Image, x, y, w, h float64) {
	if img == nil {
		return
	}
	bounds := img.Bounds()
	srcW := float64(bounds.Dx())
	srcH := float64(bounds.Dy())
	if srcW == 0 || srcH == 0 || w <= 0 || h <= 0 {
		return
	}

	scale := math.Max(w/srcW, h/srcH)
	drawW := srcW * scale
	drawH := srcH * scale
	posX := x + (w-drawW)/2.0
	posY := y + (h-drawH)/2.0

	dc.Push()
	dc.DrawRectangle(x, y, w, h)
	dc.Clip()
	dc.DrawImage(scaleImageHighQuality(img, int(math.Ceil(drawW)), int(math.Ceil(drawH))), int(posX), int(posY))
	dc.Pop()
	dc.ResetClip()
}

func drawRectangularAvatar(dc *gg.Context, x, y, w, h float64, avatarURL, initials string) {
	if strings.TrimSpace(avatarURL) != "" {
		if avatarImg, err := loadImageFromURIOrURL(avatarURL); err == nil && avatarImg != nil {
			drawImageCover(dc, avatarImg, x, y, w, h)
			return
		}
	}

	// Fallback avatar if no photo
	dc.SetColor(hexColor("#1B1B1E"))
	dc.DrawRectangle(x, y, w, h)
	dc.Fill()

	displayInitials := strings.TrimSpace(initials)
	if displayInitials == "" {
		displayInitials = "SC"
	}
	fontSize := math.Min(w, h) * 0.4
	drawText(dc, displayInitials, x+w/2.0, y+h/2.0, mustFont(fonts.HelveticaBold, fontSize), hexColor("#F6D686"), 0.5, 0.5)
}

func formatCardNumberOrRef(raw string) string {
	cleaned := strings.ReplaceAll(strings.TrimSpace(raw), "-", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "")
	if len(cleaned) == 16 {
		return fmt.Sprintf("%s %s %s %s", cleaned[0:4], cleaned[4:8], cleaned[8:12], cleaned[12:16])
	}
	if len(cleaned) >= 12 {
		return fmt.Sprintf("XXXX XXXX %s", cleaned[len(cleaned)-4:])
	}
	if cleaned == "" {
		return "XXXX XXXX 1234"
	}
	return raw
}

func formatSaintsDate(rawDate string) string {
	raw := strings.TrimSpace(rawDate)
	if raw == "" || raw == "N/A" {
		return "DD/MM/YYYY"
	}
	if strings.Contains(raw, "/") {
		return raw
	}
	formats := []string{
		"02 January 2006",
		"2 January 2006",
		"02 Jan 2006",
		"2006-01-02",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05Z",
		time.RFC3339,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, raw); err == nil {
			return t.Format("02/01/2006")
		}
	}
	return raw
}

func drawVerticalGradientLines(dc *gg.Context, lines []string, x, y, lineSpacing float64, face font.Face, topHex, bottomHex string, stop1 float64, ax, ay float64) {
	var validLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			validLines = append(validLines, l)
		}
	}
	if len(validLines) == 0 {
		return
	}
	dc.SetFontFace(face)
	maxLineW := 0.0
	for _, l := range validLines {
		w, _ := dc.MeasureString(l)
		if w > maxLineW {
			maxLineW = w
		}
	}
	if maxLineW <= 0 {
		return
	}

	fh := dc.FontHeight()
	pad := 60.0
	totalContentH := float64(len(validLines)-1)*lineSpacing + fh
	imgW := int(math.Ceil(maxLineW + pad*2))
	imgH := int(math.Ceil(totalContentH + pad*2))

	tempDC := gg.NewContext(imgW, imgH)
	tempDC.SetFontFace(face)
	tempDC.SetColor(color.White)

	for i, l := range validLines {
		curLineY := pad + float64(i)*lineSpacing + fh*0.5
		tempDC.DrawStringAnchored(l, pad, curLineY, 0, 0.5)
	}

	maskImg := tempDC.Image()
	bounds := maskImg.Bounds()

	minY, maxY := bounds.Max.Y, bounds.Min.Y
	for py := bounds.Min.Y; py < bounds.Max.Y; py++ {
		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			_, _, _, a := maskImg.At(px, py).RGBA()
			if a > 1000 {
				if py < minY {
					minY = py
				}
				if py > maxY {
					maxY = py
				}
			}
		}
	}

	if minY >= maxY {
		for i, l := range validLines {
			drawText(dc, l, x, y+float64(i)*lineSpacing, face, hexColor(topHex), ax, ay)
		}
		return
	}

	c1 := hexColor(topHex)
	c2 := hexColor(bottomHex)
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	totalH := float64(maxY - minY)
	gradImg := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	for py := minY; py <= maxY; py++ {
		t := float64(py-minY) / totalH
		var factor float64
		if t <= stop1 {
			factor = 0.0
		} else {
			factor = (t - stop1) / (1.0 - stop1)
		}
		if factor > 1.0 {
			factor = 1.0
		}

		r := uint8(math.Round(float64(r1>>8)*(1.0-factor) + float64(r2>>8)*factor))
		g := uint8(math.Round(float64(g1>>8)*(1.0-factor) + float64(g2>>8)*factor))
		b := uint8(math.Round(float64(b1>>8)*(1.0-factor) + float64(b2>>8)*factor))

		for px := bounds.Min.X; px < bounds.Max.X; px++ {
			_, _, _, a := maskImg.At(px, py).RGBA()
			if a == 0 {
				continue
			}
			alpha := uint8(a >> 8)
			gradImg.Set(px, py, color.NRGBA{
				R: r,
				G: g,
				B: b,
				A: alpha,
			})
		}
	}

	posX := x - ax*maxLineW - pad
	posY := y - pad - fh*0.5
	dc.DrawImage(gradImg, int(math.Round(posX)), int(math.Round(posY)))
}

func drawVerticalGradientText(dc *gg.Context, text string, x, y float64, face font.Face, topHex, bottomHex string, stop1 float64, ax, ay float64) {
	drawVerticalGradientLines(dc, []string{text}, x, y, 0, face, topHex, bottomHex, stop1, ax, ay)
}

func splitMemberNameToTwoLines(dc *gg.Context, name string, maxWidth float64, baseFontSize float64, sc float64) ([]string, font.Face) {
	clean := strings.TrimSpace(name)
	if clean == "" {
		face := mustFont(fonts.AptosRegular, baseFontSize*sc)
		return []string{"Aarusha Zahia"}, face
	}

	curSize := baseFontSize
	minSize := 180.0
	for curSize >= minSize {
		face := mustFont(fonts.AptosRegular, curSize*sc)
		dc.SetFontFace(face)
		w, _ := dc.MeasureString(clean)
		if w <= maxWidth {
			return []string{clean}, face
		}

		words := strings.Fields(clean)
		if len(words) <= 1 {
			curSize -= 20.0
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

		curSize -= 20.0
	}

	face := mustFont(fonts.AptosRegular, minSize*sc)
	dc.SetFontFace(face)
	words := strings.Fields(clean)
	if len(words) <= 1 {
		return []string{clean}, face
	}
	splitIdx := len(words) / 2
	return []string{strings.Join(words[:splitIdx], " "), strings.Join(words[splitIdx:], " ")}, face
}

func drawSaintsClubMemberCardFront(data employeeCardTemplateData) ([]byte, error) {
	w, h := saintsCardW, saintsCardH
	dc := gg.NewContext(w, h)

	sx := float64(w) / float64(saintsCardW)
	sy := float64(h) / float64(saintsCardH)
	sc := math.Min(sx, sy)

	// 1. Background
	if bgImg, err := loadImageFromURIOrURL(saintsBgURL); err == nil && bgImg != nil {
		dc.DrawImage(scaleImageHighQuality(bgImg, w, h), 0, 0)
	} else {
		dc.SetColor(hexColor("#0E0E12"))
		dc.DrawRectangle(0, 0, float64(w), float64(h))
		dc.Fill()
	}

	// 2. Lower Bar
	barH := 322.0 * sy
	barY := float64(h) - barH
	if barImg, err := loadImageFromURIOrURL(saintsLowerBarURL); err == nil && barImg != nil {
		croppedBar := autoCropNonEmpty(barImg)
		dc.DrawImage(scaleImageHighQuality(croppedBar, w, int(math.Ceil(barH))), 0, int(barY))
	} else {
		dc.SetColor(hexColor("#BC925C"))
		dc.DrawRectangle(0, barY, float64(w), barH)
		dc.Fill()
	}

	// 3. Top-Left: Logo & Main Titles
	logoX := 460.0 * sx
	logoY := 460.0 * sy
	logoSize := 640.0 * sc

	if logoImg, err := loadImageFromURIOrURL(saintsLogoURL); err == nil && logoImg != nil {
		croppedLogo := autoCropNonEmpty(logoImg)
		drawImageContain(dc, croppedLogo, logoX, logoY, logoSize, logoSize)
	}

	titlesX := logoX + logoSize + 160.0*sx
	titlesH := 330.0 * sc
	titlesW := titlesH * (3872.0 / 554.0)
	titlesY := logoY + (logoSize-titlesH)/2.0

	if titlesImg, err := loadImageFromURIOrURL(saintsMainTitlesURL); err == nil && titlesImg != nil {
		croppedTitles := autoCropNonEmpty(titlesImg)
		drawImageContain(dc, croppedTitles, titlesX, titlesY, titlesW, titlesH)
	} else {
		drawVerticalGradientText(dc, "SAINTS CLUB MEMBER CARD", titlesX, titlesY+titlesH*0.35, mustFont(fonts.HelveticaBold, 60.0*sc), "#C99032", "#EAE2B1", 0.202, 0, 0.5)
		drawVerticalGradientText(dc, "ONE CLUB ALLIANCE", titlesX, titlesY+titlesH*0.75, mustFont(fonts.HelveticaBold, 50.0*sc), "#C99032", "#EAE2B1", 0.202, 0, 0.5)
	}

	// 4. Right Column: Photo Frame & Avatar
	frameH := 1700.0 * sc
	frameW := frameH * (2890.0 / 3342.0)
	frameRightMargin := 460.0 * sx
	frameX := float64(w) - frameRightMargin - frameW
	frameY := 1690.0 * sy

	// Frame inner hollow ratios:
	innerXRatio := 119.0 / 2890.0
	innerYRatio := 119.0 / 3342.0
	innerWRatio := 2651.0 / 2890.0
	innerHRatio := 3103.0 / 3342.0

	photoX := frameX + frameW*innerXRatio
	photoY := frameY + frameH*innerYRatio
	photoW := frameW * innerWRatio
	photoH := frameH * innerHRatio

	drawRectangularAvatar(dc, photoX, photoY, photoW, photoH, data.AvatarURL, data.Initials)

	if frameImg, err := loadImageFromURIOrURL(saintsPhotoFrameURL); err == nil && frameImg != nil {
		croppedFrame := autoCropNonEmpty(frameImg)
		drawImageContain(dc, croppedFrame, frameX, frameY, frameW, frameH)
	} else {
		dc.SetColor(hexColor("#D4AF37"))
		dc.SetLineWidth(25.0 * sc)
		dc.DrawRectangle(photoX, photoY, photoW, photoH)
		dc.Stroke()
	}

	// 4.1. Top-Right: NFC Icon (above Photo Frame, white color)
	nfcH := 515.0 * sc
	nfcW := 410.0 * sc
	nfcX := frameX + frameW - nfcW
	nfcY := 650.0 * sy

	if nfcImg, err := loadImageFromURIOrURL(saintsNfcURL); err == nil && nfcImg != nil {
		croppedNFC := autoCropNonEmpty(nfcImg)
		tintedNFC := tintImageToColor(croppedNFC, "#FFFFFF")
		if croppedNFC.Bounds().Dy() > 0 {
			nfcW = nfcH * (float64(croppedNFC.Bounds().Dx()) / float64(croppedNFC.Bounds().Dy()))
			nfcX = frameX + frameW - nfcW
		}
		scaledNFC := scaleImageHighQuality(tintedNFC, int(math.Round(nfcW)), int(math.Round(nfcH)))
		dc.DrawImage(scaledNFC, int(math.Round(nfcX)), int(math.Round(nfcY)))
	} else if nfcFallback, err := renderNFCIcon("#FFFFFF"); err == nil && nfcFallback != nil {
		drawImageContain(dc, nfcFallback, nfcX, nfcY, nfcW, nfcH)
	}

	// 5. Center Column: Payment QR Code
	qrPanelSize := 1260.0 * sc
	qrPanelX := 2890.0 * sx
	qrPanelY := 1910.0 * sy
	qrPanelRadius := 36.0 * sc

	drawRoundedPanel(dc, qrPanelX, qrPanelY, qrPanelSize, qrPanelSize, qrPanelRadius, color.White)

	qrPadding := 35.0 * sc
	qrContentX := qrPanelX + qrPadding
	qrContentY := qrPanelY + qrPadding
	qrContentSize := qrPanelSize - qrPadding*2.0

	qrImg, err := loadImageOrGeneratePaymentQRCode(data.PaymentQRCode, 1800)
	if err == nil && qrImg != nil {
		blackQR := ensureBlackQRCodeImage(qrImg)
		drawImageContain(dc, blackQR, qrContentX, qrContentY, qrContentSize, qrContentSize)
	}

	// Bangla QR Center Logo
	banglaSize := 250.0 * sc
	banglaCenterX := qrPanelX + qrPanelSize/2.0
	banglaCenterY := qrPanelY + qrPanelSize/2.0
	banglaRadius := 20.0 * sc

	drawRoundedPanel(dc, banglaCenterX-banglaSize/2.0, banglaCenterY-banglaSize/2.0, banglaSize, banglaSize, banglaRadius, color.White)

	if banglaImg, err := loadImageFromURIOrURL(saintsBanglaQRLogoURL); err == nil && banglaImg != nil {
		drawImageContain(dc, banglaImg, banglaCenterX-banglaSize/2.0+10.0*sc, banglaCenterY-banglaSize/2.0+10.0*sc, banglaSize-20.0*sc, banglaSize-20.0*sc)
	} else {
		drawText(dc, "BANGLA", banglaCenterX, banglaCenterY-22.0*sy, mustFont(fonts.HelveticaBold, 32.0*sc), hexColor("#E31D2B"), 0.5, 0.5)
		drawText(dc, "QR", banglaCenterX, banglaCenterY+24.0*sy, mustFont(fonts.HelveticaBold, 38.0*sc), hexColor("#00A859"), 0.5, 0.5)
	}

	// 6. Left Column: Member Details
	detailsStartX := 480.0 * sx
	memberName := universityFallbackString(data.MemberName, "Aarusha Zahia")
	maxNameWidth := qrPanelX - detailsStartX - 100.0*sx

	nameLines, nameFace := splitMemberNameToTwoLines(dc, memberName, maxNameWidth, 340.0, sc)

	var nameStartY float64
	nameLineSpacing := 360.0 * sy
	if len(nameLines) == 1 {
		nameStartY = 2000.0 * sy
	} else {
		nameStartY = 1820.0 * sy
	}

	drawVerticalGradientLines(dc, nameLines, detailsStartX, nameStartY, nameLineSpacing, nameFace, "#C99032", "#EAE2B1", 0.202, 0, 0.5)

	// Underline below name
	nameBottomY := nameStartY + float64(len(nameLines)-1)*nameLineSpacing
	lineY := nameBottomY + 200.0*sy
	lineH := 70.0 * sc
	lineW := 2150.0 * sc
	if lineW > (qrPanelX - detailsStartX - 100.0*sx) {
		lineW = qrPanelX - detailsStartX - 100.0*sx
	}

	if lineImg, err := loadImageFromURIOrURL(saintsLineURL); err == nil && lineImg != nil {
		croppedLine := autoCropNonEmpty(lineImg)
		dc.DrawImage(scaleImageHighQuality(croppedLine, int(lineW), int(lineH)), int(detailsStartX), int(lineY))
	} else {
		dc.SetColor(hexColor("#BC925C"))
		dc.SetLineWidth(6.0 * sc)
		dc.DrawLine(detailsStartX, lineY+lineH/2.0, detailsStartX+lineW, lineY+lineH/2.0)
		dc.Stroke()
	}

	// Membership ID label & value (with UI gradient)
	labelFont := mustFont(fonts.AptosRegular, 240.0*sc)
	valueFont := mustFont(fonts.AptosRegular, 300.0*sc)

	label1Y := lineY + 380.0*sy
	val1Y := label1Y + 360.0*sy
	drawVerticalGradientText(dc, "Membership ID:", detailsStartX, label1Y, labelFont, "#C99032", "#EAE2B1", 0.202, 0, 0.5)

	memIDStr := formatCardNumberOrRef(universityFallbackString(data.RegNo, "XXXX XXXX 1234"))
	drawVerticalGradientText(dc, memIDStr, detailsStartX, val1Y, valueFont, "#C99032", "#EAE2B1", 0.202, 0, 0.5)

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode saints club front image: %w", err)
	}
	return out.Bytes(), nil
}

func formatSaintsAddress(rawAddr string) (string, string) {
	raw := strings.TrimSpace(rawAddr)
	if raw == "" || raw == "N/A" || strings.Contains(raw, "QCH7+4GV") || strings.Contains(raw, "QCH7+4GC") {
		return "QCH7+4GC, Road No.7,", "Dhaka 1212,"
	}
	parts := strings.Split(raw, ",")
	if len(parts) >= 3 {
		line1 := strings.TrimSpace(parts[0]) + ", " + strings.TrimSpace(parts[1]) + ","
		line2 := strings.TrimSpace(strings.Join(parts[2:], ", "))
		return line1, line2
	}
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]) + ",", strings.TrimSpace(parts[1])
	}
	return raw, ""
}

func drawSaintsClubMemberCardBack(data employeeCardTemplateData) ([]byte, error) {
	w, h := saintsCardW, saintsCardH
	dc := gg.NewContext(w, h)

	sx := float64(w) / float64(saintsCardW)
	sy := float64(h) / float64(saintsCardH)
	sc := math.Min(sx, sy)

	// 1. Background
	if bgImg, err := loadImageFromURIOrURL(saintsBackBgURL); err == nil && bgImg != nil {
		dc.DrawImage(scaleImageHighQuality(bgImg, w, h), 0, 0)
	} else if bgImg, err := loadImageFromURIOrURL(saintsBgURL); err == nil && bgImg != nil {
		dc.DrawImage(scaleImageHighQuality(bgImg, w, h), 0, 0)
	} else {
		dc.SetColor(hexColor("#0E0E12"))
		dc.DrawRectangle(0, 0, float64(w), float64(h))
		dc.Fill()
	}

	// 2. Lower Bar
	barH := 322.0 * sy
	barY := float64(h) - barH
	if barImg, err := loadImageFromURIOrURL(saintsLowerBarURL); err == nil && barImg != nil {
		croppedBar := autoCropNonEmpty(barImg)
		dc.DrawImage(scaleImageHighQuality(croppedBar, w, int(math.Ceil(barH))), 0, int(barY))
	} else {
		dc.SetColor(hexColor("#BC925C"))
		dc.DrawRectangle(0, barY, float64(w), barH)
		dc.Fill()
	}

	// 3. Top-Left: Main Titles (compact to leave breathing room)
	titlesX := 420.0 * sx
	titlesY := 520.0 * sy
	titlesH := 265.0 * sc
	titlesW := titlesH * (3872.0 / 554.0)

	if titlesImg, err := loadImageFromURIOrURL(saintsMainTitlesURL); err == nil && titlesImg != nil {
		croppedTitles := autoCropNonEmpty(titlesImg)
		drawImageContain(dc, croppedTitles, titlesX, titlesY, titlesW, titlesH)
	} else {
		drawVerticalGradientText(dc, "SAINTS CLUB MEMBER CARD", titlesX, titlesY+titlesH*0.35, mustFont(fonts.HelveticaBold, 50.0*sc), "#C99032", "#EAE2B1", 0.202, 0, 0.5)
		drawVerticalGradientText(dc, "ONE CLUB ALLIANCE", titlesX, titlesY+titlesH*0.75, mustFont(fonts.HelveticaBold, 42.0*sc), "#C99032", "#EAE2B1", 0.202, 0, 0.5)
	}

	// 4. Right Column: Large VDS Frame & QR Code
	// Aligned with the right edge, matching the exact UI proportion
	vdsSize := 3140.0 * sc
	vdsRightMargin := 260.0 * sx
	vdsX := float64(w) - vdsRightMargin - vdsSize
	vdsY := 380.0 * sy

	// Frame inner hollow ratios:
	vdsInnerRatio := 80.0 / 3123.0
	vdsInnerSizeRatio := 2964.0 / 3123.0

	qrBoxX := vdsX + vdsSize*vdsInnerRatio
	qrBoxY := vdsY + vdsSize*vdsInnerRatio
	qrBoxSize := vdsSize * vdsInnerSizeRatio

	// Fill white square inside frame
	dc.SetColor(color.White)
	dc.DrawRectangle(qrBoxX, qrBoxY, qrBoxSize, qrBoxSize)
	dc.Fill()

	qrPadding := 55.0 * sc
	vdsQRX := qrBoxX + qrPadding
	vdsQRY := qrBoxY + qrPadding
	vdsQRSize := qrBoxSize - qrPadding*2.0

	oneIDQRImg, err := loadImageOrGenerateVDSQRCode(data.OneIDQRCode, 2400)
	if err == nil && oneIDQRImg != nil {
		drawImageContain(dc, oneIDQRImg, vdsQRX, vdsQRY, vdsQRSize, vdsQRSize)
	}

	if vdsFrameImg, err := loadImageFromURIOrURL(saintsVDSFrameURL); err == nil && vdsFrameImg != nil {
		croppedVDSFrame := autoCropNonEmpty(vdsFrameImg)
		drawImageContain(dc, croppedVDSFrame, vdsX, vdsY, vdsSize, vdsSize)
	} else {
		dc.SetColor(hexColor("#D4AF37"))
		dc.SetLineWidth(25.0 * sc)
		dc.DrawRectangle(qrBoxX, qrBoxY, qrBoxSize, qrBoxSize)
		dc.Stroke()
	}

	// 5. Middle-Left: Terms & Contact Text
	termsStartX := 420.0 * sx
	termsY := 2140.0 * sy
	lineSpacing := 170.0 * sy
	bodyFont := mustFont(fonts.AptosRegular, 125.0*sc)

	termsLines := []string{
		"This card is the property of Saints club",
		"and must presented upon request. Use",
		"of this card is subject to the terms and",
		"conditions of membership.",
	}

	drawVerticalGradientLines(dc, termsLines, termsStartX, termsY, lineSpacing, bodyFont, "#C99032", "#EAE2B1", 0.202, 0, 0.5)

	// Contact info
	contactStartY := termsY + float64(len(termsLines))*lineSpacing + 130.0*sy
	phone := universityFallbackString(data.PhoneNumber, "012XXXXXXXX")
	addrLine1, addrLine2 := formatSaintsAddress(data.Address)

	contactLines := []string{
		"Tel. " + phone,
		"Address: " + addrLine1,
	}
	if addrLine2 != "" {
		contactLines = append(contactLines, addrLine2)
	}

	drawVerticalGradientLines(dc, contactLines, termsStartX, contactStartY, lineSpacing, bodyFont, "#C99032", "#EAE2B1", 0.202, 0, 0.5)

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode saints club back image: %w", err)
	}
	return out.Bytes(), nil
}

// BuildSaintsClubMemberCardImages builds both front (Payment) and back (OneID) images for Saints Club members.
func BuildSaintsClubMemberCardImages(
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
		Initials:      universityBuildInitials(universityFallbackString(employee.UserName, "Aarusha Zahia")),
		AvatarURL:     universityFallbackString(employee.AvatarURL, ""),
		OneIDQRCode:   universityFallbackImage(generateVDS.ExtractedQRImage),
		PaymentQRCode: paymentQRValue,
		IssueDate:     branding.IssueDate,
		ExpiryDate:    branding.ExpiryDate,
		BusinessName:  universityFallbackString(merchant.BusinessName, "Saints Club"),
		PhoneNumber:   universityFallbackString(merchant.Phone, "012XXXXXXXX"),
		Address:       universityFallbackString(branding.Address, "QCH7+4GC, Road No.7, Dhaka 1212,"),
	}

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
