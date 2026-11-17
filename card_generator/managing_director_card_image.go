package card_generator

import (
	"bytes"
	"employee-card-admin/assets/fonts"
	"employee-card-admin/common/constant"
	"fmt"
	"image/color"
	"image/png"
	"math"
	"strings"

	"github.com/fogleman/gg"
)

func isManagingDirectorDesignation(designation string) bool {
	return strings.EqualFold(strings.TrimSpace(designation), constant.DesignationManager)
}

func drawManagingDirectorCardFront(data employeeCardTemplateData) ([]byte, error) {
	w, h := cardFrontW, cardFrontH
	dc := gg.NewContext(w, h)
	sx := float64(w) / 306.0
	sy := float64(h) / 485.0
	sc := math.Min(sx, sy)

	dc.SetColor(hexColor("#292672"))
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()

	for _, sq := range []struct{ x, y, wh, alpha float64 }{
		{25, 200, 24, .28}, {62, 210, 34, .40}, {8, 235, 26, .24},
		{73, 175, 30, .50}, {83, 175, 30, .20}, {-6, 233, 26, .24},
		{42, 250, 30, .55}, {-10, 290, 28, .50}, {-1, 324, 24, .30},
		{50, 340, 32, .60}, {-26, 333, 40, .50}, {65, 380, 26, .45},
		{55, 435, 22, .50}, {195, 145, 40, .50}, {210, 145, 40, .20},
		{165, 175, 30, .50}, {177, 175, 30, .20}, {247, 270, 34, .40},
		{247, 308, 34, .80}, {298, 355, 34, .80}, {283, 377, 50, .40},
	} {
		dc.SetColor(withAlpha(hexColor("#0D6EFD"), sq.alpha))
		dc.DrawRoundedRectangle(sq.x*sx, sq.y*sy, sq.wh*sx, sq.wh*sy, 3*sc)
		dc.Fill()
	}

	headerX, headerY := 30*sx, 16*sy
	headerW, headerH := 248*sx, 80*sy
	drawRoundedPanel(dc, headerX, headerY, headerW, headerH, 12*sc, color.White)
	drawAvatarImage(dc, headerX+54*sx, headerY+60*sy, 42*sc, data.AvatarURL, data.Initials)

	logoSize := 42 * sc
	uniCenterX := headerX + 166*sx
	logoX := uniCenterX - logoSize/2
	logoY := headerY + 8*sy
	if logoImg, err := loadImageFromURIOrURL(data.LOGO); err == nil && logoImg != nil {
		drawImageContain(dc, logoImg, logoX, logoY, logoSize, logoSize)
	}
	drawText(dc, strings.Title(strings.ToLower(data.BusinessName)), uniCenterX, headerY+58*sy, mustFont(fonts.AptosRegular, 10.5*sc), hexColor("#1A1A2E"), 0.5, 0.5)

	drawText(dc, strings.ToUpper(data.MemberName), 30*sx, 136*sy, mustFont(fonts.AptosBold, 15*sc), color.White, 0, 0.5)
	drawText(dc, "Managing Director, Dhaka Bank PLC.", 30*sx, 158*sy, mustFont(fonts.AptosRegular, 12*sc), color.White, 0, 0.5)
	drawText(dc, "Sponsor", 30*sx, 176*sy, mustFont(fonts.AptosRegular, 10*sc), color.White, 0, 0.5)

	dc.Push()
	dc.Translate(286*sx, 350*sy)
	dc.Rotate(-math.Pi / 2)
	drawText(
		dc,
		"ID - "+data.OneID,
		0+30*sx,
		0,
		mustFont(fonts.AptosRegular, 9*sc),
		color.White,
		0.5,
		0.5,
	)
	dc.Pop()

	infoX, infoY := 30*sx, 190*sy
	infoW, infoH := 242*sx, 268*sy
	drawRoundedPanel(dc, infoX, infoY, infoW, infoH, 8*sc, color.White)

	badgeX := infoX + infoW - 48*sx
	badgeY := infoY - 14*sy

	bloodImg, err := renderBloodIcon("#ff0000")
	if err == nil {
		drawImageContain(
			dc,
			bloodImg,
			badgeX,
			badgeY,
			40*sx,
			40*sy,
		)
	}

	drawText(
		dc,
		data.BloodGroup,
		badgeX+28*sx,
		badgeY+9*sy,
		mustFont(fonts.AptosBold, 5*sc),
		color.White,
		0.5,
		0.5,
	)

	drawText(dc, "Personal Information", infoX+infoW/2, infoY+20*sy, mustFont(fonts.AptosBold, 13*sc), hexColor("#5E1EB2"), 0.5, 0.5)
	if qrImg, err := loadImageOrGenerateVDSQRCode(data.OneIDQRCode, 1800); err == nil && qrImg != nil {
		drawImageContain(dc, qrImg, infoX+12*sx, infoY+38*sy, 218*sx, 218*sy)
	}

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode managing director front image: %w", err)
	}
	return out.Bytes(), nil
}

func drawManagingDirectorCardBack(data employeeCardTemplateData) ([]byte, error) {
	w, h := cardBackW, cardBackH
	dc := gg.NewContext(w, h)
	sx := float64(w) / 459.0
	sy := float64(h) / 727.92
	sc := math.Min(sx, sy)

	dc.SetColor(hexColor("#292672"))
	dc.DrawRectangle(0, 0, float64(w), float64(h))
	dc.Fill()
	drawDotStrip(dc, w, 0, int(33*sy))

	panelX, panelY := 47*sx, 60*sy
	drawRoundedPanel(dc, panelX, panelY, 365*sx, 240*sy, 10*sc, color.White)
	if qrImg, err := loadImageOrGeneratePaymentQRCode(data.PaymentQRCode, 1800); err == nil && qrImg != nil {
		drawImageContain(dc, qrImg, panelX+24*sx, panelY+24*sy, 200*sx, 200*sy)
	}
	drawText(dc, "Payment", panelX+302*sx, panelY+37.5*sy, mustFont(fonts.AptosBold, 22.5*sc), hexColor("#631FB4"), 0.5, 0.5)
	if nfcImg, err := renderNFCIcon("#631FB4"); err == nil {
		drawImageContain(dc, nfcImg, panelX+263*sx, panelY+86*sy, 78*sx, 78*sy)
	}

	labelX, colonX, valueX := 47*sx, 155*sx, 179*sx
	issueY := 340.5 * sy
	drawMDDateRow(dc, "Issue Date", data.IssueDate, labelX, colonX, valueX, issueY, sc)
	drawMDDateRow(dc, "Expiry Date", data.ExpiryDate, labelX, colonX, valueX, issueY+27*sy, sc)

	uniY := 424.5 * sy
	drawText(dc, strings.Title(strings.ToLower(data.BusinessName)), 47*sx, uniY, mustFont(fonts.AptosBold, 22.68*sc), hexColor("#C7A4FF"), 0, 0.5)
	drawMDPhoneIcon(dc, 57.5*sx, 466*sy, 9*sc)
	drawText(dc, "02-41010626", 78*sx, 462*sy, mustFont(fonts.AptosBold, 16.26*sc), color.White, 0, 0.5)
	drawMDLocationIcon(dc, 56*sx, 503*sy, 7.5*sc)
	dc.SetFontFace(mustFont(fonts.AptosRegular, 13.35*sc))
	dc.SetColor(color.White)
	dc.DrawStringWrapped(data.Address, 75*sx, 497*sy, 0, 0.5, 305*sx, 1.15, gg.AlignLeft)

	sigY, sigH := 559.92*sy, 69*sy
	dc.SetColor(color.White)
	dc.DrawRectangle(0, sigY, float64(w), sigH)
	dc.Fill()
	sigImg, err := loadImageFromURIOrURL("https://res.cloudinary.com/llhcc8la/image/upload/v1784778889/student/1784778888-Pasted%20image.png.png")
	if err == nil && sigImg != nil {
		drawImageContain(dc, sigImg, 20*sx, sigY+5*sy, 190*sx, 65*sy)
	}
	drawText(dc, "Authorised Signature", 260*sx, sigY+34.5*sy, mustFont(fonts.AptosBold, 13.95*sc), hexColor("#0B0B0B"), 0, 0.5)

	drawText(dc, "This ID card is valid for authorized person only,", float64(w)/2, 655.92*sy, mustFont(fonts.AptosItalic, 15*sc), color.White, 0.5, 0.5)
	drawText(dc, "If lost & found, please contact the above address", float64(w)/2, 673.92*sy, mustFont(fonts.AptosItalic, 15*sc), color.White, 0.5, 0.5)
	drawDotStrip(dc, w, h-int(33*sy), int(33*sy))

	var out bytes.Buffer
	if err := png.Encode(&out, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode managing director back image: %w", err)
	}
	return out.Bytes(), nil
}

func drawMDDateRow(dc *gg.Context, label, value string, labelX, colonX, valueX, y, sc float64) {
	drawText(dc, label, labelX, y, mustFont(fonts.AptosRegular, 18.9*sc), hexColor("#C7A4FF"), 0, 0.5)
	drawText(dc, ":", colonX, y, mustFont(fonts.AptosRegular, 18.9*sc), color.White, 0, 0.5)
	drawText(dc, value, valueX, y, mustFont(fonts.AptosRegular, 18.9*sc), color.White, 0, 0.5)
}

func drawMDPhoneIcon(dc *gg.Context, x, y, r float64) {
	dc.SetColor(hexColor("#C7A4FF"))
	scale := (2.0 * r) / 7.0
	dc.MoveTo(x+(1.5402-3.5)*scale, y+(0.0200043-4.0)*scale)
	dc.CubicTo(x+(1.6902-3.5)*scale, y+(0.0600043-4.0)*scale, x+(1.8102-3.5)*scale, y+(0.149999-4.0)*scale, x+(1.8802-3.5)*scale, y+(0.279999-4.0)*scale)
	dc.CubicTo(x+(1.9202-3.5)*scale, y+(0.349999-4.0)*scale, x+(1.9502-3.5)*scale, y+(0.420006-4.0)*scale, x+(1.9802-3.5)*scale, y+(0.490005-4.0)*scale)
	dc.CubicTo(x+(2.1102-3.5)*scale, y+(0.830006-4.0)*scale, x+(2.2102-3.5)*scale, y+(1.17-4.0)*scale, x+(2.2702-3.5)*scale, y+(1.53-4.0)*scale)
	dc.CubicTo(x+(2.3202-3.5)*scale, y+(1.8-4.0)*scale, x+(2.2302-3.5)*scale, y+(1.97001-4.0)*scale, x+(1.9602-3.5)*scale, y+(2.10001-4.0)*scale)
	dc.CubicTo(x+(1.7702-3.5)*scale, y+(2.19001-4.0)*scale, x+(1.5702-3.5)*scale, y+(2.27001-4.0)*scale, x+(1.3802-3.5)*scale, y+(2.35001-4.0)*scale)
	dc.CubicTo(x+(1.2802-3.5)*scale, y+(2.39001-4.0)*scale, x+(1.25021-3.5)*scale, y+(2.46001-4.0)*scale, x+(1.28021-3.5)*scale, y+(2.57001-4.0)*scale)
	dc.CubicTo(x+(1.28021-3.5)*scale, y+(2.59001-4.0)*scale, x+(1.2902-3.5)*scale, y+(2.62-4.0)*scale, x+(1.3002-3.5)*scale, y+(2.64-4.0)*scale)
	dc.CubicTo(x+(1.5502-3.5)*scale, y+(3.35-4.0)*scale, x+(1.9202-3.5)*scale, y+(4.00001-4.0)*scale, x+(2.3802-3.5)*scale, y+(4.59001-4.0)*scale)
	dc.CubicTo(x+(2.6002-3.5)*scale, y+(4.87001-4.0)*scale, x+(2.8402-3.5)*scale, y+(5.14001-4.0)*scale, x+(3.1002-3.5)*scale, y+(5.38-4.0)*scale)
	dc.CubicTo(x+(3.1802-3.5)*scale, y+(5.46-4.0)*scale, x+(3.2802-3.5)*scale, y+(5.45-4.0)*scale, x+(3.3502-3.5)*scale, y+(5.36-4.0)*scale)
	dc.CubicTo(x+(3.4802-3.5)*scale, y+(5.19-4.0)*scale, x+(3.63021-3.5)*scale, y+(5.03-4.0)*scale, x+(3.81021-3.5)*scale, y+(4.91-4.0)*scale)
	dc.CubicTo(x+(3.90021-3.5)*scale, y+(4.85-4.0)*scale, x+(4.0002-3.5)*scale, y+(4.81-4.0)*scale, x+(4.1002-3.5)*scale, y+(4.8-4.0)*scale)
	dc.CubicTo(x+(4.2202-3.5)*scale, y+(4.79-4.0)*scale, x+(4.3302-3.5)*scale, y+(4.83-4.0)*scale, x+(4.4302-3.5)*scale, y+(4.89-4.0)*scale)
	dc.CubicTo(x+(4.7602-3.5)*scale, y+(5.1-4.0)*scale, x+(5.0702-3.5)*scale, y+(5.34-4.0)*scale, x+(5.33019-3.5)*scale, y+(5.63-4.0)*scale)
	dc.CubicTo(x+(5.39019-3.5)*scale, y+(5.70001-4.0)*scale, x+(5.45021-3.5)*scale, y+(5.78001-4.0)*scale, x+(5.50021-3.5)*scale, y+(5.87001-4.0)*scale)
	dc.CubicTo(x+(5.60021-3.5)*scale, y+(6.05001-4.0)*scale, x+(5.5802-3.5)*scale, y+(6.23001-4.0)*scale, x+(5.4602-3.5)*scale, y+(6.40001-4.0)*scale)
	dc.CubicTo(x+(5.3002-3.5)*scale, y+(6.64001-4.0)*scale, x+(5.08019-3.5)*scale, y+(6.8-4.0)*scale, x+(4.83019-3.5)*scale, y+(6.91-4.0)*scale)
	dc.CubicTo(x+(4.70019-3.5)*scale, y+(6.96-4.0)*scale, x+(4.5602-3.5)*scale, y+(6.99001-4.0)*scale, x+(4.4302-3.5)*scale, y+(7.04001-4.0)*scale)
	dc.CubicTo(x+(4.4302-3.5)*scale, y+(7.04001-4.0)*scale, x+(4.43021-3.5)*scale, y+(7.04001-4.0)*scale, x+(4.42021-3.5)*scale, y+(7.04001-4.0)*scale)
	dc.LineTo(x+(3.8502-3.5)*scale, y+(7.04001-4.0)*scale)
	dc.LineTo(x+(3.83019-3.5)*scale, y+(7.04001-4.0)*scale)
	dc.CubicTo(x+(3.63019-3.5)*scale, y+(7.00001-4.0)*scale, x+(3.4402-3.5)*scale, y+(6.94001-4.0)*scale, x+(3.2602-3.5)*scale, y+(6.85001-4.0)*scale)
	dc.CubicTo(x+(2.7202-3.5)*scale, y+(6.58001-4.0)*scale, x+(2.2402-3.5)*scale, y+(6.22-4.0)*scale, x+(1.8202-3.5)*scale, y+(5.78-4.0)*scale)
	dc.CubicTo(x+(1.3902-3.5)*scale, y+(5.33-4.0)*scale, x+(1.0302-3.5)*scale, y+(4.82-4.0)*scale, x+(0.730204-3.5)*scale, y+(4.28-4.0)*scale)
	dc.CubicTo(x+(0.420204-3.5)*scale, y+(3.72-4.0)*scale, x+(0.200195-3.5)*scale, y+(3.13-4.0)*scale, x+(0.0801951-3.5)*scale, y+(2.5-4.0)*scale)
	dc.CubicTo(x+(0.0201951-3.5)*scale, y+(2.18-4.0)*scale, x+(-0.019797-3.5)*scale, y+(1.85-4.0)*scale, x+(0.010203-3.5)*scale, y+(1.52-4.0)*scale)
	dc.CubicTo(x+(0.050203-3.5)*scale, y+(1.06-4.0)*scale, x+(0.250205-3.5)*scale, y+(0.689999-4.0)*scale, x+(0.590205-3.5)*scale, y+(0.389999-4.0)*scale)
	dc.CubicTo(x+(0.800205-3.5)*scale, y+(0.209999-4.0)*scale, x+(1.03021-3.5)*scale, y+(0.07-4.0)*scale, x+(1.31021-3.5)*scale, y+(0.0-4.0)*scale)
	dc.LineTo(x+(1.56021-3.5)*scale, y+(0.0-4.0)*scale)
	dc.LineTo(x+(1.5402-3.5)*scale, y+(0.0200043-4.0)*scale)
	dc.ClosePath()
	dc.Fill()
	dc.MoveTo(x+(6.48022-3.5)*scale, y+(3.28-4.0)*scale)
	dc.CubicTo(x+(6.48022-3.5)*scale, y+(3.5-4.0)*scale, x+(6.46022-3.5)*scale, y+(3.71001-4.0)*scale, x+(6.42022-3.5)*scale, y+(3.93001-4.0)*scale)
	dc.CubicTo(x+(6.39022-3.5)*scale, y+(4.08001-4.0)*scale, x+(6.26022-3.5)*scale, y+(4.18001-4.0)*scale, x+(6.12022-3.5)*scale, y+(4.18001-4.0)*scale)
	dc.CubicTo(x+(5.96022-3.5)*scale, y+(4.18001-4.0)*scale, x+(5.83021-3.5)*scale, y+(4.09-4.0)*scale, x+(5.79021-3.5)*scale, y+(3.94-4.0)*scale)
	dc.CubicTo(x+(5.77021-3.5)*scale, y+(3.86-4.0)*scale, x+(5.79021-3.5)*scale, y+(3.78001-4.0)*scale, x+(5.80021-3.5)*scale, y+(3.70001-4.0)*scale)
	dc.CubicTo(x+(5.91021-3.5)*scale, y+(3.02001-4.0)*scale, x+(5.77022-3.5)*scale, y+(2.39001-4.0)*scale, x+(5.37022-3.5)*scale, y+(1.82001-4.0)*scale)
	dc.CubicTo(x+(5.05022-3.5)*scale, y+(1.36001-4.0)*scale, x+(4.63021-3.5)*scale, y+(1.04-4.0)*scale, x+(4.10021-3.5)*scale, y+(0.860002-4.0)*scale)
	dc.CubicTo(x+(4.00021-3.5)*scale, y+(0.820002-4.0)*scale, x+(3.89022-3.5)*scale, y+(0.800006-4.0)*scale, x+(3.78022-3.5)*scale, y+(0.770006-4.0)*scale)
	dc.CubicTo(x+(3.57022-3.5)*scale, y+(0.720006-4.0)*scale, x+(3.47021-3.5)*scale, y+(0.500004-4.0)*scale, x+(3.57021-3.5)*scale, y+(0.300004-4.0)*scale)
	dc.CubicTo(x+(3.63021-3.5)*scale, y+(0.180004-4.0)*scale, x+(3.77021-3.5)*scale, y+(0.110001-4.0)*scale, x+(3.91021-3.5)*scale, y+(0.140001-4.0)*scale)
	dc.CubicTo(x+(4.99021-3.5)*scale, y+(0.380001-4.0)*scale, x+(5.75021-3.5)*scale, y+(1.01001-4.0)*scale, x+(6.21021-3.5)*scale, y+(2.01001-4.0)*scale)
	dc.CubicTo(x+(6.37021-3.5)*scale, y+(2.36001-4.0)*scale, x+(6.45021-3.5)*scale, y+(2.73-4.0)*scale, x+(6.47021-3.5)*scale, y+(3.11-4.0)*scale)
	dc.CubicTo(x+(6.47021-3.5)*scale, y+(3.17-4.0)*scale, x+(6.47021-3.5)*scale, y+(3.22-4.0)*scale, x+(6.47021-3.5)*scale, y+(3.28-4.0)*scale)
	dc.LineTo(x+(6.48022-3.5)*scale, y+(3.28-4.0)*scale)
	dc.ClosePath()
	dc.Fill()
	dc.MoveTo(x+(5.4302-3.5)*scale, y+(3.37-4.0)*scale)
	dc.CubicTo(x+(5.4302-3.5)*scale, y+(3.54-4.0)*scale, x+(5.41021-3.5)*scale, y+(3.69999-4.0)*scale, x+(5.37021-3.5)*scale, y+(3.85999-4.0)*scale)
	dc.CubicTo(x+(5.33021-3.5)*scale, y+(4.01999-4.0)*scale, x+(5.16021-3.5)*scale, y+(4.12001-4.0)*scale, x+(5.00021-3.5)*scale, y+(4.09001-4.0)*scale)
	dc.CubicTo(x+(4.83021-3.5)*scale, y+(4.05001-4.0)*scale, x+(4.73021-3.5)*scale, y+(3.90001-4.0)*scale, x+(4.75021-3.5)*scale, y+(3.73001-4.0)*scale)
	dc.CubicTo(x+(4.77021-3.5)*scale, y+(3.60001-4.0)*scale, x+(4.79021-3.5)*scale, y+(3.47-4.0)*scale, x+(4.78021-3.5)*scale, y+(3.33-4.0)*scale)
	dc.CubicTo(x+(4.75021-3.5)*scale, y+(2.69-4.0)*scale, x+(4.45021-3.5)*scale, y+(2.22-4.0)*scale, x+(3.89021-3.5)*scale, y+(1.91-4.0)*scale)
	dc.CubicTo(x+(3.76021-3.5)*scale, y+(1.84-4.0)*scale, x+(3.6102-3.5)*scale, y+(1.79-4.0)*scale, x+(3.4602-3.5)*scale, y+(1.76-4.0)*scale)
	dc.CubicTo(x+(3.3302-3.5)*scale, y+(1.73-4.0)*scale, x+(3.2502-3.5)*scale, y+(1.64999-4.0)*scale, x+(3.2202-3.5)*scale, y+(1.52999-4.0)*scale)
	dc.CubicTo(x+(3.1602-3.5)*scale, y+(1.29999-4.0)*scale, x+(3.3502-3.5)*scale, y+(1.08999-4.0)*scale, x+(3.6002-3.5)*scale, y+(1.13999-4.0)*scale)
	dc.CubicTo(x+(3.8802-3.5)*scale, y+(1.18999-4.0)*scale, x+(4.1402-3.5)*scale, y+(1.3-4.0)*scale, x+(4.3802-3.5)*scale, y+(1.46-4.0)*scale)
	dc.CubicTo(x+(4.9402-3.5)*scale, y+(1.83-4.0)*scale, x+(5.28021-3.5)*scale, y+(2.34999-4.0)*scale, x+(5.40021-3.5)*scale, y+(2.99999-4.0)*scale)
	dc.CubicTo(x+(5.42021-3.5)*scale, y+(3.11999-4.0)*scale, x+(5.4302-3.5)*scale, y+(3.25-4.0)*scale, x+(5.4402-3.5)*scale, y+(3.38-4.0)*scale)
	dc.LineTo(x+(5.4302-3.5)*scale, y+(3.37-4.0)*scale)
	dc.ClosePath()
	dc.Fill()
	dc.MoveTo(x+(4.36021-3.5)*scale, y+(3.49-4.0)*scale)
	dc.CubicTo(x+(4.35021-3.5)*scale, y+(3.57-4.0)*scale, x+(4.3402-3.5)*scale, y+(3.67-4.0)*scale, x+(4.3202-3.5)*scale, y+(3.77-4.0)*scale)
	dc.CubicTo(x+(4.2802-3.5)*scale, y+(3.93-4.0)*scale, x+(4.1102-3.5)*scale, y+(4.03-4.0)*scale, x+(3.9602-3.5)*scale, y+(4.0-4.0)*scale)
	dc.CubicTo(x+(3.7902-3.5)*scale, y+(3.96-4.0)*scale, x+(3.6702-3.5)*scale, y+(3.81-4.0)*scale, x+(3.7102-3.5)*scale, y+(3.64-4.0)*scale)
	dc.CubicTo(x+(3.7602-3.5)*scale, y+(3.39-4.0)*scale, x+(3.71021-3.5)*scale, y+(3.17001-4.0)*scale, x+(3.53021-3.5)*scale, y+(2.98001-4.0)*scale)
	dc.CubicTo(x+(3.43021-3.5)*scale, y+(2.87001-4.0)*scale, x+(3.3002-3.5)*scale, y+(2.79001-4.0)*scale, x+(3.1502-3.5)*scale, y+(2.76001-4.0)*scale)
	dc.CubicTo(x+(2.9302-3.5)*scale, y+(2.71001-4.0)*scale, x+(2.82021-3.5)*scale, y+(2.49-4.0)*scale, x+(2.92021-3.5)*scale, y+(2.29-4.0)*scale)
	dc.CubicTo(x+(2.99021-3.5)*scale, y+(2.16-4.0)*scale, x+(3.13021-3.5)*scale, y+(2.1-4.0)*scale, x+(3.28021-3.5)*scale, y+(2.13-4.0)*scale)
	dc.CubicTo(x+(3.83021-3.5)*scale, y+(2.24-4.0)*scale, x+(4.2702-3.5)*scale, y+(2.71001-4.0)*scale, x+(4.3502-3.5)*scale, y+(3.26001-4.0)*scale)
	dc.CubicTo(x+(4.3502-3.5)*scale, y+(3.33001-4.0)*scale, x+(4.3602-3.5)*scale, y+(3.40001-4.0)*scale, x+(4.3702-3.5)*scale, y+(3.48001-4.0)*scale)
	dc.LineTo(x+(4.36021-3.5)*scale, y+(3.49-4.0)*scale)
	dc.ClosePath()
	dc.Fill()

}

func drawMDLocationIcon(dc *gg.Context, x, y, r float64) {
	dc.SetColor(hexColor("#C7A4FF"))
	scale := (2.0 * r) / 6.0
	dc.MoveTo(x+(2.71806-3.0)*scale, y+(8.32679-4.5)*scale)
	dc.CubicTo(x+(2.64806-3.0)*scale, y+(8.20679-4.5)*scale, x+(2.58805-3.0)*scale, y+(8.09679-4.5)*scale, x+(2.52805-3.0)*scale, y+(7.99679-4.5)*scale)
	dc.CubicTo(x+(1.95805-3.0)*scale, y+(7.03679-4.5)*scale, x+(1.38806-3.0)*scale, y+(6.0768-4.5)*scale, x+(0.828056-3.0)*scale, y+(5.1168-4.5)*scale)
	dc.CubicTo(x+(0.508056-3.0)*scale, y+(4.5668-4.5)*scale, x+(0.24805-3.0)*scale, y+(3.97679-4.5)*scale, x+(0.0880505-3.0)*scale, y+(3.35679-4.5)*scale)
	dc.CubicTo(x+(-0.14195-3.0)*scale, y+(2.46679-4.5)*scale, x+(0.0880523-3.0)*scale, y+(1.6868-4.5)*scale, x+(0.668052-3.0)*scale, y+(1.0068-4.5)*scale)
	dc.CubicTo(x+(2.03805-3.0)*scale, y+(-0.583205-4.5)*scale, x+(4.51806-3.0)*scale, y+(-0.233211-4.5)*scale, x+(5.40806-3.0)*scale, y+(1.65679-4.5)*scale)
	dc.CubicTo(x+(5.76806-3.0)*scale, y+(2.43679-4.5)*scale, x+(5.71804-3.0)*scale, y+(3.20679-4.5)*scale, x+(5.34804-3.0)*scale, y+(3.95679-4.5)*scale)
	dc.CubicTo(x+(5.17804-3.0)*scale, y+(4.29679-4.5)*scale, x+(5.00805-3.0)*scale, y+(4.63679-4.5)*scale, x+(4.80805-3.0)*scale, y+(4.95679-4.5)*scale)
	dc.CubicTo(x+(4.14805-3.0)*scale, y+(6.02679-4.5)*scale, x+(3.48805-3.0)*scale, y+(7.09679-4.5)*scale, x+(2.81805-3.0)*scale, y+(8.15679-4.5)*scale)
	dc.CubicTo(x+(2.78805-3.0)*scale, y+(8.20679-4.5)*scale, x+(2.75806-3.0)*scale, y+(8.24679-4.5)*scale, x+(2.71806-3.0)*scale, y+(8.31679-4.5)*scale)
	dc.LineTo(x+(2.71806-3.0)*scale, y+(8.32679-4.5)*scale)
	dc.ClosePath()
	dc.MoveTo(x+(4.42805-3.0)*scale, y+(2.73679-4.5)*scale)
	dc.CubicTo(x+(4.42805-3.0)*scale, y+(1.85679-4.5)*scale, x+(3.69805-3.0)*scale, y+(1.1268-4.5)*scale, x+(2.81805-3.0)*scale, y+(1.1168-4.5)*scale)
	dc.CubicTo(x+(1.92805-3.0)*scale, y+(1.1168-4.5)*scale, x+(1.19805-3.0)*scale, y+(1.85679-4.5)*scale, x+(1.19805-3.0)*scale, y+(2.74679-4.5)*scale)
	dc.CubicTo(x+(1.19805-3.0)*scale, y+(3.62679-4.5)*scale, x+(1.92805-3.0)*scale, y+(4.35679-4.5)*scale, x+(2.80805-3.0)*scale, y+(4.35679-4.5)*scale)
	dc.CubicTo(x+(3.69805-3.0)*scale, y+(4.35679-4.5)*scale, x+(4.42805-3.0)*scale, y+(3.6168-4.5)*scale, x+(4.42805-3.0)*scale, y+(2.7268-4.5)*scale)
	dc.LineTo(x+(4.42805-3.0)*scale, y+(2.73679-4.5)*scale)
	dc.ClosePath()
	dc.Fill()

}
