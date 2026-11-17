package card_generator

import (
	"bytes"
	"employee-card-admin/assets/fonts"
	"employee-card-admin/common/constant"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"strings"

	"github.com/fogleman/gg"
)

const (
	facultyCardW               = 3942 // 2.19 in at 1800 DPI
	facultyCardH               = 6174 // 3.43 in at 1800 DPI 6174
	facultyBgURL               = "UAP_Faculty_Card_UI/1790673773-Backgound-(Gradient--with-Earth)-@4x.png"
	facultyNfcURL              = "UAP_Faculty_Card_UI/1790673833-NFC.png"
	facultyLeftParenthesisURL  = "UAP_Faculty_Card_UI/1790673917-Parentheses-Symbol-Right.png"
	facultyRightParenthesisURL = "UAP_Faculty_Card_UI/1790673877-Parentheses-Symbol-Left.png"
	facultyBloodDropURL        = "UAP_Faculty_Card_UI/1790673961-Blood-Group.png"
	facultyJointlyIssuedURL    = "UAP_Faculty_Card_UI/1790674001-Jointly-Issued-by.png"
	facultyWhiteLineURL        = "UAP_Faculty_Card_UI/1790674032-White-Line.png"
)

func isFacultyDesignation(designation string) bool {
	d := strings.ToLower(strings.TrimSpace(designation))
	faculty := strings.ToLower(constant.DesignationFaculty)
	return d == faculty || strings.Contains(d, faculty) ||
		d == "falculty" || strings.Contains(d, "falculty") ||
		strings.Contains(d, "professor") || strings.Contains(d, "lecturer") || strings.Contains(d, "teacher")
}

func tintImageToColor(src image.Image, targetHex string) image.Image {
	if src == nil {
		return nil
	}
	targetClr := hexColor(targetHex)
	tR, tG, tB, _ := targetClr.RGBA()
	targetR := uint8(tR >> 8)
	targetG := uint8(tG >> 8)
	targetB := uint8(tB >> 8)

	bounds := src.Bounds()
	dst := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := src.At(x, y)
			_, _, _, a := c.RGBA()
			if a == 0 {
				continue
			}
			alpha := uint8(a >> 8)
			dst.Set(x, y, color.NRGBA{
				R: targetR,
				G: targetG,
				B: targetB,
				A: alpha,
			})
		}
	}
	return dst
}

func drawFacultyCardFront(data employeeCardTemplateData) ([]byte, error) {
	w, h := facultyCardW, facultyCardH
	dc := gg.NewContext(w, h)

	sx := float64(w) / 401.0
	sy := float64(h) / 637.0
	sc := math.Min(sx, sy)

	if bgImg, err := loadImageFromURIOrURL(facultyBgURL); err == nil && bgImg != nil {
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

	if parenthesisImg, err := loadImageFromURIOrURL(facultyLeftParenthesisURL); err == nil && parenthesisImg != nil {
		tintedImg := tintImageToColor(parenthesisImg, "#38D9C0")
		drawImageContain(dc, tintedImg, badgeRightX, badgeRightY, badgeRightW, badgeRightH)
	} else {
		dc.DrawRoundedRectangle(badgeRightX, badgeRightY, badgeRightW, badgeRightH, badgeRightR)
		dc.SetColor(withAlpha(hexColor("#150A36"), 0.35))
		dc.FillPreserve()
		dc.SetColor(hexColor("#38D9C0"))
		dc.SetLineWidth(3.5 * sc)
		dc.Stroke()
	}

	dc.Push()
	dc.Translate(badgeRightX+badgeRightW/2+6.5*sc, badgeRightY+badgeRightH/2)
	dc.Rotate(-math.Pi / 2)
	drawText(dc, "FACULTY", 0, 0, mustFont(fonts.HelveticaBold, 23.5*sc), color.White, 0.5, 0.5)
	dc.Pop()

	// badgeLeftX := 24.0 * sx
	// badgeLeftY := 165.0 * sy
	// badgeLeftW := 54.0 * sx
	// badgeLeftH := 220.0 * sy
	// badgeLeftR := 26.0 * sc

	// if parenthesisImg, err := loadImageFromURIOrURL(facultyRightParenthesisURL); err == nil && parenthesisImg != nil {
	// 	tintedImg := tintImageToColor(parenthesisImg, "#38D9C0")
	// 	drawImageContain(dc, tintedImg, badgeLeftX, badgeLeftY, badgeLeftW, badgeLeftH)
	// } else {
	// 	dc.DrawRoundedRectangle(badgeLeftX, badgeLeftY, badgeLeftW, badgeLeftH, badgeLeftR)
	// 	dc.SetColor(withAlpha(hexColor("#150A36"), 0.35))
	// 	dc.FillPreserve()
	// 	dc.SetColor(hexColor("#38D9C0"))
	// 	dc.SetLineWidth(3.5 * sc)
	// 	dc.Stroke()
	// }

	// regNoStr := universityFallbackString(data.RegNo, "00000000")
	// regLabelFont := mustFont(fonts.HelveticaBold, 14.0*sc)
	// regValFont := mustFont(fonts.HelveticaBold, 20.0*sc)

	// dc.SetFontFace(regLabelFont)
	// regLabelW, _ := dc.MeasureString("REG NO : ")
	// dc.SetFontFace(regValFont)
	// regValW, _ := dc.MeasureString(regNoStr)
	// totalRegW := regLabelW + regValW

	// dc.Push()
	// dc.Translate(badgeLeftX+badgeLeftW/2-15*sc, badgeLeftY+badgeLeftH/2)
	// dc.Rotate(-math.Pi / 2)
	// startRegX := -totalRegW / 2.0
	// drawText(dc, "REG NO : ", startRegX+regLabelW/2, 0, regLabelFont, hexColor("#38D9C0"), 0.5, 0.5)
	// drawText(dc, regNoStr, startRegX+regLabelW+regValW/2, 0, regValFont, color.White, 0.5, 0.5)
	// dc.Pop()

	avatarCenterX := 200.5 * sx
	avatarCenterY := 252.0 * sy
	avatarR := 95.0 * sc

	drawAvatarImage(dc, avatarCenterX, avatarCenterY, avatarR, data.AvatarURL, data.Initials)

	dc.DrawCircle(avatarCenterX, avatarCenterY, avatarR)
	dc.SetColor(color.White)
	dc.SetLineWidth(1.5 * sc)
	dc.Stroke()

	memberName := strings.ToUpper(universityFallbackString(data.MemberName, "MEHREEN AFREE TUSHTI"))

	maxNameWidth := 235.0 * sx

	facultyDesignation := strings.ToUpper(strings.TrimSpace(data.Designation))
	if facultyDesignation == "" || facultyDesignation == "N/A" || facultyDesignation == "FACULTY" {
		facultyDesignation = "PROFESSOR"
	}

	rawDept := strings.TrimSpace(data.Department)
	if rawDept == "" || rawDept == "N/A" {
		rawDept = "EEE"
	}
	deptText := "DEPARTMENT : " + strings.ToUpper(rawDept)

	nameFont := mustFont(fonts.HelveticaBold, 20.0*sc)
	dc.SetFontFace(nameFont)
	nameLines := splitNameToLines(dc, memberName, maxNameWidth)

	var desigY, deptY float64
	if len(nameLines) >= 2 {
		drawText(dc, nameLines[0], avatarCenterX, 365.0*sy, nameFont, color.White, 0.5, 0.5)
		drawText(dc, nameLines[1], avatarCenterX, 386.0*sy, nameFont, color.White, 0.5, 0.5)
		desigY = 407.0 * sy
		deptY = 428.0 * sy
	} else {
		singleNameFont := mustFont(fonts.HelveticaBold, 21.5*sc)
		drawText(dc, memberName, avatarCenterX, 372.0*sy, singleNameFont, color.White, 0.5, 0.5)
		desigY = 396.0 * sy
		deptY = 420.0 * sy
	}

	desigFont := mustFont(fonts.HelveticaBold, 18.5*sc)
	dc.SetFontFace(desigFont)
	if desigW, _ := dc.MeasureString(facultyDesignation); desigW > maxNameWidth {
		desigFont = mustFont(fonts.HelveticaBold, 18.5*sc*(maxNameWidth/desigW))
	}
	drawText(dc, facultyDesignation, avatarCenterX, desigY, desigFont, hexColor("#38D9C0"), 0.5, 0.5)

	deptFont := mustFont(fonts.HelveticaBold, 17.5*sc)
	dc.SetFontFace(deptFont)
	if deptW, _ := dc.MeasureString(deptText); deptW > maxNameWidth {
		deptFont = mustFont(fonts.HelveticaBold, 17.5*sc*(maxNameWidth/deptW))
	}
	drawText(dc, deptText, avatarCenterX, deptY, deptFont, hexColor("#38D9C0"), 0.5, 0.5)

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
	if nfcImg, err := loadImageFromURIOrURL(facultyNfcURL); err == nil && nfcImg != nil {
		tintedImg := tintImageToColor(nfcImg, "#38D9C0")
		drawImageContain(dc, tintedImg, nfcX, nfcY, nfcW, nfcH)
	}

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode faculty front image: %w", err)
	}
	return out.Bytes(), nil
}

func drawFacultyCardBack(data employeeCardTemplateData) ([]byte, error) {
	w, h := facultyCardW, facultyCardH
	dc := gg.NewContext(w, h)

	sx := float64(w) / 401.0
	sy := float64(h) / 637.0
	sc := math.Min(sx, sy)

	if bgImg, err := loadImageFromURIOrURL(facultyBgURL); err == nil && bgImg != nil {
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

	if dropImg, err := loadImageFromURIOrURL(facultyBloodDropURL); err == nil && dropImg != nil {
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

	if jointlyImg, err := loadImageFromURIOrURL(facultyJointlyIssuedURL); err == nil && jointlyImg != nil {
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
	if lineImg, err := loadImageFromURIOrURL(facultyWhiteLineURL); err == nil && lineImg != nil {
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
		return nil, fmt.Errorf("encode faculty back image: %w", err)
	}
	return out.Bytes(), nil
}
