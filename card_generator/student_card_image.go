package card_generator

import (
	"bytes"
	"employee-card-admin/assets/fonts"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"

	"github.com/fogleman/gg"
	"golang.org/x/image/draw"
)

const (
	studentCardW               = 3942 // 2.19 in at 1800 DPI
	studentCardH               = 6174 // 3.43 in at 1800 DPI
	studentBgURL               = "UAP_Student_Card_UI/1790672133-Backgound-(Gradient--with-Earth)-.png"
	studentNfcURL              = "UAP_Student_Card_UI/1790672287-NFC.png"
	studentLeftParenthesisURL  = "UAP_Student_Card_UI/1790673193-Parentheses-Symbol-Right.png"
	studentRightParenthesisURL = "UAP_Student_Card_UI/1790672348-Parentheses-Symbol-Left.png"
	studentBloodDropURL        = "UAP_Student_Card_UI/1790672418-Blood-Group.png"
	studentJointlyIssuedURL    = "UAP_Student_Card_UI/1790672494-Jointly-Issued-by.png"
	studentWhiteLineURL        = "UAP_Student_Card_UI/1790672600-White-Line.png"
	banglaQRLogoURL            = "UAP_Student_Card_UI/1790672994-Bangla-QR.png"
)

func isStudentDesignation(designation string) bool {
	d := strings.ToLower(strings.TrimSpace(designation))
	return d == "student" || strings.Contains(d, "student")
}

func scaleImageHighQuality(src image.Image, width, height int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW == 0 || srcH == 0 || width <= 0 || height <= 0 {
		return src
	}

	if srcW == width && srcH == height {
		return src
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}

func drawStudentCardFront(data employeeCardTemplateData) ([]byte, error) {
	w, h := studentCardW, studentCardH
	dc := gg.NewContext(w, h)

	sx := float64(w) / 401.0
	sy := float64(h) / 637.0
	sc := math.Min(sx, sy)

	if bgImg, err := loadImageFromURIOrURL(studentBgURL); err == nil && bgImg != nil {
		dc.DrawImage(scaleImageHighQuality(bgImg, w, h), 0, 0)
	} else {
		dc.SetColor(hexColor("#1F104F"))
		dc.DrawRectangle(0, 0, float64(w), float64(h))
		dc.Fill()
	}

	logoW := 180.0 * sx
	logoH := 220.0 * sy
	logoX := (float64(w) - logoW) / 2.0
	logoY := 10.0*sy - 35*sc
	if logoImg, err := loadImageFromURIOrURL(data.LOGO); err == nil && logoImg != nil {
		drawImageContain(dc, logoImg, logoX, logoY, logoW, logoH)
	}

	badgeRightX := 322.0 * sx
	badgeRightY := 5.0 * sy
	badgeRightW := 54.0 * sx
	badgeRightH := 230.0 * sy
	badgeRightR := 40.0 * sc

	if parenthesisImg, err := loadImageFromURIOrURL(studentLeftParenthesisURL); err == nil && parenthesisImg != nil {
		drawImageContain(dc, parenthesisImg, badgeRightX, badgeRightY, badgeRightW, badgeRightH)
	} else {
		dc.DrawRoundedRectangle(badgeRightX, badgeRightY, badgeRightW, badgeRightH, badgeRightR)
		dc.SetColor(withAlpha(hexColor("#150A36"), 0.35))
		dc.FillPreserve()
		dc.SetColor(hexColor("#D81B80"))
		dc.SetLineWidth(3.5 * sc)
		dc.Stroke()
	}

	dc.Push()
	dc.Translate(badgeRightX+badgeRightW/2+6.5*sc, badgeRightY+badgeRightH/2)
	dc.Rotate(-math.Pi / 2)
	drawText(dc, "STUDENT", 0, 0, mustFont(fonts.HelveticaBold, 23.5*sc), color.White, 0.5, 0.5)
	dc.Pop()

	badgeLeftX := 15.0 * sx
	badgeLeftY := 150.0 * sy
	badgeLeftW := 80.0 * sx
	badgeLeftH := 220.0 * sy
	badgeLeftR := 26.0 * sc

	if parenthesisImg, err := loadImageFromURIOrURL(studentRightParenthesisURL); err == nil && parenthesisImg != nil {
		drawImageContain(dc, parenthesisImg, badgeLeftX, badgeLeftY, badgeLeftW, badgeLeftH)
	} else {
		dc.DrawRoundedRectangle(badgeLeftX, badgeLeftY, badgeLeftW, badgeLeftH, badgeLeftR)
		dc.SetColor(withAlpha(hexColor("#150A36"), 0.35))
		dc.FillPreserve()
		dc.SetColor(hexColor("#D81B80"))
		dc.SetLineWidth(3.5 * sc)
		dc.Stroke()
	}

	regNoStr := universityFallbackString(data.RegNo, "00000000")
	regLabelFont := mustFont(fonts.HelveticaBold, 14.0*sc)
	regValFont := mustFont(fonts.HelveticaBold, 20.0*sc)

	dc.SetFontFace(regLabelFont)
	regLabelW, _ := dc.MeasureString("REG NO : ")
	dc.SetFontFace(regValFont)
	regValW, _ := dc.MeasureString(regNoStr)
	totalRegW := regLabelW + regValW

	dc.Push()
	dc.Translate(badgeLeftX+badgeLeftW/2-15*sc, badgeLeftY+badgeLeftH/2)
	dc.Rotate(-math.Pi / 2)
	startRegX := -totalRegW / 2.0
	drawText(dc, "REG NO : ", startRegX+regLabelW/2, 0, regLabelFont, hexColor("#38D9C0"), 0.5, 0.5)
	drawText(dc, regNoStr, startRegX+regLabelW+regValW/2, 0, regValFont, color.White, 0.5, 0.5)
	dc.Pop()

	avatarCenterX := 200.5 * sx
	avatarCenterY := 252.0 * sy
	avatarR := 95.0 * sc

	drawAvatarImage(dc, avatarCenterX, avatarCenterY, avatarR, data.AvatarURL, data.Initials)

	dc.DrawCircle(avatarCenterX, avatarCenterY, avatarR)
	dc.SetColor(color.White)
	dc.SetLineWidth(1.5 * sc)
	dc.Stroke()

	memberName := strings.ToUpper(universityFallbackString(data.MemberName, "MEHREEN AFREE TUSHTI"))

	nameFont := mustFont(fonts.HelveticaBold, 21.0*sc)
	dc.SetFontFace(nameFont)

	maxNameWidth := 235.0 * sx

	nameLines := splitNameToLines(dc, memberName, maxNameWidth)
	deptY := 425.0 * sy

	if len(nameLines) == 2 {
		drawText(dc, nameLines[0], avatarCenterX, 368.0*sy, nameFont, color.White, 0.5, 0.5)
		drawText(dc, nameLines[1], avatarCenterX, 392.0*sy, nameFont, color.White, 0.5, 0.5)
		deptY = 420.0 * sy
	} else {
		drawText(
			dc,
			memberName,
			avatarCenterX,
			390.0*sy,
			mustFont(fonts.HelveticaBold, 22.5*sc),
			color.White,
			0.5,
			0.5,
		)
	}

	rawDept := "EEE"
	deptText := "DEPARTMENT : " + strings.ToUpper(rawDept)
	drawText(dc, deptText, avatarCenterX, deptY, mustFont(fonts.HelveticaRegular, 17.0*sc), hexColor("#38D9C0"), 0.5, 0.5)

	qrPanelW := 146.0 * sx
	qrPanelH := 146.0 * sy
	qrPanelX := (float64(w) - qrPanelW) / 2.0
	qrPanelY := 446.0 * sy
	qrPanelR := 15.0 * sc

	drawRoundedPanel(dc, qrPanelX, qrPanelY, qrPanelW, qrPanelH, qrPanelR, color.White)

	qrImg, err := loadImageOrGeneratePaymentQRCode(data.PaymentQRCode, 1800)
	if err == nil && qrImg != nil {
		drawImageContain(dc, qrImg, qrPanelX+6*sx, qrPanelY+6*sy, qrPanelW-12*sx, qrPanelH-12*sy)
	}

	banglaQrCenterX := qrPanelX + qrPanelW/2
	banglaQrCenterY := qrPanelY + qrPanelH/2
	bW, bH := 36.0*sx, 36.0*sy
	drawRoundedPanel(dc, banglaQrCenterX-bW/2, banglaQrCenterY-bH/2, bW, bH, 4*sc, color.White)
	if banglaImg, err := loadImageFromURIOrURL(banglaQRLogoURL); err == nil && banglaImg != nil {
		drawImageContain(dc, banglaImg, banglaQrCenterX-bW/2+2*sx, banglaQrCenterY-bH/2+2*sy, bW-4*sx, bH-4*sy)
	} else {
		dc.DrawRoundedRectangle(banglaQrCenterX-bW/2, banglaQrCenterY-bH/2, bW, bH, 4*sc)
		dc.SetColor(hexColor("#00A859"))
		dc.SetLineWidth(1.5 * sc)
		dc.Stroke()
		drawText(dc, "BANGLA", banglaQrCenterX, banglaQrCenterY-5*sy, mustFont(fonts.HelveticaBold, 5.5*sc), hexColor("#E31D2B"), 0.5, 0.5)
		drawText(dc, "QR", banglaQrCenterX, banglaQrCenterY+6*sy, mustFont(fonts.HelveticaBold, 6.5*sc), hexColor("#00A859"), 0.5, 0.5)
	}

	nfcW := 140.0 * sx
	nfcH := 140.0 * sy
	nfcX := 350.0*sx - nfcW/2
	nfcY := (qrPanelY + qrPanelH/2) - nfcH/2
	if nfcImg, err := loadImageFromURIOrURL(studentNfcURL); err == nil && nfcImg != nil {
		drawImageContain(dc, nfcImg, nfcX, nfcY, nfcW, nfcH)
	}

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode student front image: %w", err)
	}
	return out.Bytes(), nil
}

func drawSignatureImage(dc *gg.Context, signatureURL string, x, y, w, h float64) bool {
	if strings.TrimSpace(signatureURL) == "" {
		return false
	}
	sigImg, err := loadImageFromURIOrURL(signatureURL)
	if err != nil || sigImg == nil {
		return false
	}
	drawImageContain(dc, sigImg, x, y, w, h)
	return true
}

func drawStudentCardBack(data employeeCardTemplateData) ([]byte, error) {
	w, h := studentCardW, studentCardH
	dc := gg.NewContext(w, h)

	sx := float64(w) / 401.0
	sy := float64(h) / 637.0
	sc := math.Min(sx, sy)

	if bgImg, err := loadImageFromURIOrURL(studentBgURL); err == nil && bgImg != nil {
		dc.Push()
		dc.Translate(float64(w)/2, float64(h)/2)
		dc.Rotate(math.Pi)
		dc.Translate(-float64(w)/2, -float64(h)/2)
		dc.DrawImage(scaleImageHighQuality(bgImg, w, h), 0, 0)
		dc.Pop()
	} else {
		dc.SetColor(hexColor("#1F104F"))
		dc.DrawRectangle(0, 0, float64(w), float64(h))
		dc.Fill()
	}

	qrPanelX := 56.0 * sx
	qrPanelY := 34.0 * sy
	qrPanelW := 294.0 * sx
	qrPanelH := 294.0 * sy

	oneIDQRImg, err := loadImageOrGenerateVDSQRCode(data.OneIDQRCode, 2400)
	if err == nil && oneIDQRImg != nil {
		drawImageContain(dc, oneIDQRImg, qrPanelX, qrPanelY, qrPanelW, qrPanelH)
	}

	tabX := qrPanelX + qrPanelW - 6*sx
	tabY := qrPanelY + 0*sy
	tabW := 25.0 * sx
	tabH := 190.0 * sy

	dc.SetColor(color.White)
	dc.DrawRoundedRectangle(tabX, tabY, tabW, tabH, 5*sc)
	dc.Fill()

	dc.Push()
	dc.Translate(tabX+tabW/2, tabY+tabH/2)
	dc.Rotate(math.Pi / 2)

	regularFace := mustFont(fonts.HelveticaRegular, 14*sc)
	boldFace := mustFont(fonts.HelveticaBold, 19*sc)

	seg1 := "Scan with the "
	seg2 := "VeriQR"
	seg3 := " App"

	dc.SetFontFace(regularFace)
	w1, _ := dc.MeasureString(seg1)
	dc.SetFontFace(boldFace)
	w2, _ := dc.MeasureString(seg2)
	dc.SetFontFace(regularFace)
	w3, _ := dc.MeasureString(seg3)

	totalW := w1 + w2 + w3
	startX := -totalW / 2.0

	drawText(dc, seg1, startX, 0, regularFace, color.Black, 0, 0.5)
	drawText(dc, seg2, startX+w1, 0, boldFace, color.Black, 0, 0.5)
	drawText(dc, seg3, startX+w1+w2, 0, regularFace, color.Black, 0, 0.5)
	dc.Pop()

	centerX := float64(w) / 2.0
	legalCenterX := 196 * sx
	drawText(dc, "NID-Verified Credential. Digitally", legalCenterX, 348.0*sy, mustFont(fonts.HelveticaItalic, 16*sc), color.White, 0.5, 0.5)
	drawText(dc, "Signed by Certifying Authority under", legalCenterX, 368.0*sy, mustFont(fonts.HelveticaItalic, 16*sc), color.White, 0.5, 0.5)
	drawText(dc, "the ICT Act, 2006. Legally Valid.", legalCenterX, 388.0*sy, mustFont(fonts.HelveticaItalic, 16*sc), color.White, 0.5, 0.5)

	bloodDropX := 316 * sx
	bloodDropY := 315.0 * sy
	bloodDropW := 77.0 * sx
	bloodDropH := 104 * sy

	if dropImg, err := loadImageFromURIOrURL(studentBloodDropURL); err == nil && dropImg != nil {
		drawImageContain(dc, dropImg, bloodDropX, bloodDropY, bloodDropW, bloodDropH)
	} else {
		drawRoundedPanel(dc, bloodDropX, bloodDropY+15*sy, bloodDropW, 40*sy, 10*sc, hexColor("#D81B80"))
	}
	bloodGrp := universityFallbackString(data.BloodGroup, "AB+")
	drawText(dc, bloodGrp, bloodDropX+bloodDropW/2, bloodDropY+bloodDropH*0.50, mustFont(fonts.HelveticaBold, 15.0*sc), color.White, 0.5, 0.5)

	jointlyW := 340.0 * sx
	jointlyH := 65.0 * sy
	jointlyX := (float64(w) - jointlyW) / 2.0
	jointlyY := 400.0 * sy

	if jointlyImg, err := loadImageFromURIOrURL(studentJointlyIssuedURL); err == nil && jointlyImg != nil {
		drawImageContain(dc, jointlyImg, jointlyX, jointlyY, jointlyW, jointlyH)
	}

	uniName := universityFallbackString(data.BusinessName, "University of Asia Pacific")
	drawText(dc, uniName, centerX, 478*sy, mustFont(fonts.HelveticaBold, 19.5*sc), color.White, 0.5, 0.5)

	uniAddress := "74/A Green Road, Dhaka-1215"
	drawText(dc, uniAddress, centerX, 500*sy, mustFont(fonts.HelveticaRegular, 16.0*sc), color.White, 0.5, 0.5)

	phoneLabel := "Phone: "
	phoneNum := universityFallbackString(data.PhoneNumber, "+8802-41010626-9")
	phoneRegFont := mustFont(fonts.HelveticaRegular, 16*sc)
	phoneBoldFont := mustFont(fonts.HelveticaBold, 16*sc)
	dc.SetFontFace(phoneRegFont)
	pW1, _ := dc.MeasureString(phoneLabel)
	dc.SetFontFace(phoneBoldFont)
	pW2, _ := dc.MeasureString(phoneNum)
	pTotalW := pW1 + pW2
	pStartX := centerX - pTotalW/2.0
	drawText(dc, phoneLabel, pStartX, 518.0*sy, phoneRegFont, color.White, 0, 0.5)
	drawText(dc, phoneNum, pStartX+pW1, 518.0*sy, phoneBoldFont, color.White, 0, 0.5)

	emailText := "Email: " + universityFallbackString(data.Email, "registrar@uap-bd.edu")
	drawText(dc, emailText, centerX, 535.0*sy, mustFont(fonts.HelveticaRegular, 16.0*sc), color.White, 0.5, 0.5)

	webText := "Website: " + universityFallbackString(data.WebsiteUrl, "www.uap-bd.edu")
	drawText(dc, webText, centerX, 555.0*sy, mustFont(fonts.HelveticaRegular, 16.0*sc), color.White, 0.5, 0.5)

	lineW := 410.0 * sx
	lineH := 55.0 * sy
	lineX := (float64(w) - lineW) / 2.0
	lineY := 571.0*sy - lineH/2.0
	if lineImg, err := loadImageFromURIOrURL(studentWhiteLineURL); err == nil && lineImg != nil {
		dc.DrawImage(scaleImageHighQuality(lineImg, int(lineW), int(lineH)), int(lineX), int(lineY))
	} else {
		dc.SetColor(withAlpha(color.White, 0.4))
		dc.SetLineWidth(1.2 * sc)
		dc.DrawLine(35*sx, 558*sy, 365*sx, 558*sy)
		dc.Stroke()
	}

	drawText(dc, "For OneCard related inquiries or assistance,", centerX, 582.0*sy, mustFont(fonts.HelveticaItalic, 14.5*sc), color.White, 0.5, 0.5)

	foot1 := "please contact "
	foot2 := "+880 9606 501 231"
	f1Font := mustFont(fonts.HelveticaItalic, 14.5*sc)
	f2Font := mustFont(fonts.HelveticaBoldItalic, 16.0*sc)
	dc.SetFontFace(f1Font)
	fW1, _ := dc.MeasureString(foot1)
	dc.SetFontFace(f2Font)
	fW2, _ := dc.MeasureString(foot2)
	fTotalW := fW1 + fW2
	fStartX := centerX - fTotalW/2.0
	drawText(dc, foot1, fStartX, 600.0*sy, f1Font, color.White, 0, 0.5)
	drawText(dc, foot2, fStartX+fW1, 600.0*sy, f2Font, color.White, 0, 0.5)

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode student back image: %w", err)
	}
	return out.Bytes(), nil
}

func splitNameToLines(dc *gg.Context, name string, maxWidth float64) []string {
	words := strings.Fields(name)
	if len(words) <= 1 {
		return words
	}

	measureWidth := func(text string) float64 {
		w, _ := dc.MeasureString(text)
		return w
	}

	if measureWidth(name) <= maxWidth {
		return []string{name}
	}

	if len(words) >= 5 && strings.Contains(words[0], ".") {
		splitIdx := 3
		if len(words[1]) <= 2 && strings.Contains(words[1], ".") && measureWidth(strings.Join(words[:4], " ")) <= maxWidth {
			splitIdx = 4
		}
		firstLine := strings.Join(words[:splitIdx], " ")
		secondLine := strings.Join(words[splitIdx:], " ")
		if strings.TrimSpace(secondLine) != "" && measureWidth(secondLine) <= maxWidth {
			return []string{firstLine, secondLine}
		}
	}

	bestSplit := 1
	bestDiff := math.MaxFloat64

	for i := 1; i < len(words); i++ {
		if i == 1 && len(words[0]) <= 3 && !strings.Contains(words[0], ".") && len(words) > 2 {
			continue
		}

		line1 := strings.Join(words[:i], " ")
		line2 := strings.Join(words[i:], " ")

		width1 := measureWidth(line1)
		width2 := measureWidth(line2)
		if width1 > maxWidth || width2 > maxWidth {
			continue
		}

		diff := math.Abs(width1 - width2)
		if diff < bestDiff {
			bestDiff = diff
			bestSplit = i
		}
	}

	if bestSplit == 1 && len(words) > 1 && len(words[0]) <= 3 && !strings.Contains(words[0], ".") && len(words) > 2 {
		line1 := strings.Join(words[:2], " ")
		line2 := strings.Join(words[2:], " ")
		width1 := measureWidth(line1)
		width2 := measureWidth(line2)
		if width1 <= maxWidth && width2 <= maxWidth {
			return []string{line1, line2}
		}
	}

	if bestSplit == 1 {
		return []string{name}
	}

	return []string{
		strings.Join(words[:bestSplit], " "),
		strings.Join(words[bestSplit:], " "),
	}
}
