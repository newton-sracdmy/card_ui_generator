package card_generator

import (
	"bytes"
	"employee-card-admin/assets/fonts"
	"employee-card-admin/dto"
	"employee-card-admin/models"
	"image"
	"os"
	"testing"
	"time"

	"github.com/fogleman/gg"
	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	_ = godotenv.Load("../.env.dev", "../.env.uat", "../.env")
	os.Exit(m.Run())
}

func TestSplitNameToLinesHandlesShortPrefixNames(t *testing.T) {
	dc := gg.NewContext(4000, 1800)
	dc.SetFontFace(mustFont(fonts.HelveticaBold, 204.0))
	maxWidth := 2300.0
	name := "MD. SHAHRIYER ISLAM SAGOR CHOWDHURY"

	lines := splitNameToLines(dc, name, maxWidth)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines for %q, got %#v", name, lines)
	}

	if got, want := lines[0], "MD. SHAHRIYER ISLAM"; got != want {
		t.Fatalf("first line mismatch: got %q want %q", got, want)
	}
	if got, want := lines[1], "SAGOR CHOWDHURY"; got != want {
		t.Fatalf("second line mismatch: got %q want %q", got, want)
	}
}

func TestStudentCardUIGeneration(t *testing.T) {
	emp := models.Employee{
		UserName:  "Md. Shahriyer Islam Sagor Chowdhury",
		OneID:     "1234567890",
		AvatarURL: "https://res.cloudinary.com/llhcc8la/image/upload/v1786441014/Card_UI/1786441013-newton-photo%20%281%29.jpg.jpg",
	}

	mem := models.Membership{
		Designation: "Student",
		RefID:       "25108046",
		MetaData: models.JSONMap{
			"program":     "Bachelor of Science in Electrical and Electronic Engineering",
			"department":  "EEE",
			"blood_group": "B+",
		},
	}

	now := time.Now()
	nextYear := now.AddDate(1, 0, 0)
	merchant := models.Merchant{
		BusinessName: "University of Asia Pacific",
		Phone:        "+8802-58157091-4, +8802-58157096",
		Email:        "registrar@uap-bd.edu",
		Settings: &models.MerchantSetting{
			BrandColor:     "#631FB4",
			Address:        "74/A Green Road, Dhaka-1215",
			Logo:           "http://res.cloudinary.com/dz1rfos2a/image/upload/v1782311040/merchant-logo/1782311038-University_of_Asia_Pacific_%28Bangladesh%29_logo%20%281%29.png.png",
			CreatedAt:      now,
			CardExpiryTime: nextYear,
		},
	}

	dgepayCard := dto.GenericResponse{
		Data: map[string]interface{}{
			"user_emp_cards_data": map[string]interface{}{
				"card_number": "4123 4567 8901 2345",
				"qr_code":     "dummy-payment-qr-code-data",
				"merchant_info": map[string]interface{}{
					"business_name": "DG Infotech",
				},
				"user_info": map[string]interface{}{
					"phone_number": "+880198765432",
				},
			},
		},
	}

	generateVDS := dto.GenerateVDSData{
		ExtractedQRImage: "",
	}

	images, err := BuildUniversityCardImages(emp, mem, merchant, dgepayCard, generateVDS)
	if err != nil {
		t.Fatalf("Failed to build student images: %v", err)
	}

	if err := os.WriteFile("test_student_front.png", images.Payment, 0644); err != nil {
		t.Fatalf("Failed to write student front card: %v", err)
	}

	if err := os.WriteFile("test_student_back.png", images.OneID, 0644); err != nil {
		t.Fatalf("Failed to write student back card: %v", err)
	}

	t.Log("Successfully generated test_student_front.png and test_student_back.png in card_generator directory")
}

func TestUniversityCardUIGeneration(t *testing.T) {
	emp := models.Employee{
		UserName:  "Al Jabir Newton",
		OneID:     "1234567890",
		AvatarURL: "https://res.cloudinary.com/dz1rfos2a/image/upload/v1785219804/test_student/1785219802-Screenshot%20from%202026-07-28%2012-22-28.png.png",
	}

	mem := models.Membership{
		Designation: "Software Engineer",
		RefID:       "REG-9999",
		MetaData: models.JSONMap{
			"program":     "Bachelor of Science in Electrical and Electronic Engineering",
			"department":  "Electrical and Electronic Engineering",
			"blood_group": "A+",
		},
	}

	now := time.Now()
	nextYear := now.AddDate(1, 0, 0)
	merchant := models.Merchant{
		BusinessName: "University of Asia Pacific",
		Phone:        "+8802-58157091-4, +8802-58157096",
		Email:        "registrar@uap-bd.edu",
		Settings: &models.MerchantSetting{
			BrandColor:     "#631FB4",
			Address:        "74/A, Green Road, Farmgate Dhaka-1215",
			CreatedAt:      now,
			CardExpiryTime: nextYear,
		},
	}

	dgepayCard := dto.GenericResponse{
		Data: map[string]interface{}{
			"user_emp_cards_data": map[string]interface{}{
				"card_number": "4123 4567 8901 2345",
				"qr_code":     "dummy-payment-qr-code-data",
				"merchant_info": map[string]interface{}{
					"business_name": "DG Infotech",
				},
				"user_info": map[string]interface{}{
					"phone_number": "+880198765432",
				},
			},
		},
	}

	generateVDS := dto.GenerateVDSData{
		ExtractedQRImage: "",
	}

	images, err := BuildUniversityCardImages(emp, mem, merchant, dgepayCard, generateVDS)
	if err != nil {
		t.Fatalf("Failed to build images: %v", err)
	}

	err = os.WriteFile("test_payment_card.png", images.Payment, 0644)
	if err != nil {
		t.Fatalf("Failed to write payment card: %v", err)
	}

	err = os.WriteFile("test_oneid_card.png", images.OneID, 0644)
	if err != nil {
		t.Fatalf("Failed to write front card: %v", err)
	}

	t.Log("Successfully generated test_payment_card.png and test_oneid_card.png in card_generator directory")
}

func TestManagingDirectorCardUIGeneration(t *testing.T) {
	emp := models.Employee{
		UserName:  "Rashed Mahmud",
		OneID:     "9327880931",
		AvatarURL: "https://res.cloudinary.com/dz1rfos2a/image/upload/v1785219804/test_student/1785219802-Screenshot%20from%202026-07-28%2012-22-28.png.png",
	}

	mem := models.Membership{
		Designation: "Managing Director, Dhaka Bank PLC",
		RefID:       "MD-001",
		MetaData: models.JSONMap{
			"blood_group": "AB+",
		},
	}

	now := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	merchant := models.Merchant{
		BusinessName: "University of Asia Pacific",
		Phone:        "02-41010626",
		Email:        "registrar@uap-bd.edu",
		Settings: &models.MerchantSetting{
			BrandColor:     "#631FB4",
			Address:        "74/A Green Rd, Dhaka 1205",
			Logo:           "http://res.cloudinary.com/dz1rfos2a/image/upload/v1782311040/merchant-logo/1782311038-University_of_Asia_Pacific_%28Bangladesh%29_logo%20%281%29.png.png",
			CreatedAt:      now,
			CardExpiryTime: now.AddDate(5, 0, 0),
		},
	}

	dgepayCard := dto.GenericResponse{
		Data: map[string]interface{}{
			"user_emp_cards_data": map[string]interface{}{
				"card_number": "4123 4567 8901 2345",
				"qr_code":     "dummy-payment-qr-code-data",
				"merchant_info": map[string]interface{}{
					"business_name": "DG Infotech",
				},
			},
		},
	}

	images, err := BuildUniversityCardImages(emp, mem, merchant, dgepayCard, dto.GenerateVDSData{})
	if err != nil {
		t.Fatalf("Failed to build managing director images: %v", err)
	}
	if len(images.OneID) == 0 || len(images.Payment) == 0 {
		t.Fatal("Expected managing director front and back images")
	}

	if err := os.WriteFile("test_managing_director_front.png", images.OneID, 0644); err != nil {
		t.Fatalf("Failed to write managing director front card: %v", err)
	}
	if err := os.WriteFile("test_managing_director_back.png", images.Payment, 0644); err != nil {
		t.Fatalf("Failed to write managing director back card: %v", err)
	}

	t.Log("Successfully generated test_managing_director_front.png and test_managing_director_back.png in card_generator directory")
}

func TestFacultyCardUIGeneration(t *testing.T) {
	emp := models.Employee{
		UserName:  "Dr. G. R. Ahmed Jamal",
		OneID:     "1234567890",
		AvatarURL: "https://res.cloudinary.com/llhcc8la/image/upload/v1786441014/Card_UI/1786441013-newton-photo%20%281%29.jpg.jpg",
	}

	mem := models.Membership{
		Designation: "Professor",
		RefID:       "84596659",
		MetaData: models.JSONMap{
			"program":     "Bachelor of Science in Electrical and Electronic Engineering",
			"department":  "EEE",
			"designation": "Professor",
			"blood_group": "A+",
		},
	}

	now := time.Now()
	nextYear := now.AddDate(1, 0, 0)
	merchant := models.Merchant{
		BusinessName: "University of Asia Pacific",
		Phone:        "+8802-58157091-4, +8802-58157096",
		Email:        "registrar@uap-bd.edu",
		Settings: &models.MerchantSetting{
			BrandColor:     "#631FB4",
			Address:        "74/A Green Road, Dhaka-1215",
			Logo:           "http://res.cloudinary.com/dz1rfos2a/image/upload/v1782311040/merchant-logo/1782311038-University_of_Asia_Pacific_%28Bangladesh%29_logo%20%281%29.png.png",
			CreatedAt:      now,
			CardExpiryTime: nextYear,
		},
	}

	dgepayCard := dto.GenericResponse{
		Data: map[string]interface{}{
			"user_emp_cards_data": map[string]interface{}{
				"card_number": "4123 4567 8901 2345",
				"qr_code":     "dummy-payment-qr-code-data",
				"merchant_info": map[string]interface{}{
					"business_name": "DG Infotech",
				},
				"user_info": map[string]interface{}{
					"phone_number": "+880198765432",
				},
			},
		},
	}

	generateVDS := dto.GenerateVDSData{
		ExtractedQRImage: "",
	}

	images, err := BuildUniversityCardImages(emp, mem, merchant, dgepayCard, generateVDS)
	if err != nil {
		t.Fatalf("Failed to build faculty images: %v", err)
	}

	if err := os.WriteFile("test_faculty_front.png", images.Payment, 0644); err != nil {
		t.Fatalf("Failed to write faculty front card: %v", err)
	}

	if err := os.WriteFile("test_faculty_back.png", images.OneID, 0644); err != nil {
		t.Fatalf("Failed to write faculty back card: %v", err)
	}

	t.Log("Successfully generated test_faculty_front.png and test_faculty_back.png in card_generator directory")
}

func TestSaintsClubMemberCardUIGeneration(t *testing.T) {
	emp := models.Employee{
		UserName:  "Al Jabir Newton ",
		OneID:     "1234567890",
		AvatarURL: "SAINTS_CLUB_MEMBER_UI/1791274464-newton-photo-(1).jpg",
	}

	mem := models.Membership{
		Designation: "Saints Club Member",
		RefID:       "XXXX XXXX 1234",
	}

	now := time.Now()
	nextYear := now.AddDate(1, 0, 0)

	avatarImg, avatarErr := loadImageFromURIOrURL(emp.AvatarURL)
	t.Logf("Avatar load result: err=%v, nil=%v", avatarErr, avatarImg == nil)
	merchant := models.Merchant{
		BusinessName: "Saints Club",
		Phone:        "012XXXXXXXX",
		Settings: &models.MerchantSetting{
			BrandColor:     "#F6D686",
			Address:        "QCH7+4GC, Road No.7, Dhaka 1212,",
			CreatedAt:      now,
			CardExpiryTime: nextYear,
		},
	}

	dgepayCard := dto.GenericResponse{
		Data: map[string]interface{}{
			"user_emp_cards_data": map[string]interface{}{
				"card_number": "XXXX XXXX 1234",
				"qr_code":     "dummy-payment-qr-code-data",
				"merchant_info": map[string]interface{}{
					"business_name": "Saints Club",
				},
				"user_info": map[string]interface{}{
					"phone_number": "012XXXXXXXX",
				},
			},
		},
	}

	generateVDS := dto.GenerateVDSData{
		ExtractedQRImage: "",
	}

	images, err := BuildUniversityCardImages(emp, mem, merchant, dgepayCard, generateVDS)
	if err != nil {
		t.Fatalf("Failed to build saints club card images: %v", err)
	}

	if len(images.Payment) == 0 {
		t.Fatal("Expected non-empty Payment (front) card bytes")
	}
	if len(images.OneID) == 0 {
		t.Fatal("Expected non-empty OneID (back) card bytes")
	}

	if err := os.WriteFile("test_saints_front.png", images.Payment, 0644); err != nil {
		t.Fatalf("Failed to write saints front card: %v", err)
	}

	if err := os.WriteFile("test_saints_back.png", images.OneID, 0644); err != nil {
		t.Fatalf("Failed to write saints back card: %v", err)
	}

	// Also generate with a longer name to test 2nd line wrapping
	empLong := emp
	empLong.UserName = "Al Jabir Newton Chowdhury"
	imagesLong, err := BuildUniversityCardImages(empLong, mem, merchant, dgepayCard, generateVDS)
	if err == nil && len(imagesLong.Payment) > 0 {
		_ = os.WriteFile("test_saints_front_long.png", imagesLong.Payment, 0644)
	}

	t.Log("Successfully generated test_saints_front.png, test_saints_front_long.png, and test_saints_back.png")
}

func TestSaintsClubMemberThumbnailGeneration(t *testing.T) {
	emp := models.Employee{
		ID:        "saints-thumb-test-emp-1",
		UserName:  "Aarusha Zahia",
		AvatarURL: "SAINTS_CLUB_MEMBER_UI/1791274464-newton-photo-(1).jpg",
		Phone:     "+8801700000000",
		Email:     "aarusha@saintsclub.com",
		OneID:     "123456789",
	}

	mem := models.Membership{
		ID:          "saints-thumb-test-mem-1",
		RefID:       "1234-5678-9012-3456",
		Designation: "Saints Club Member",
		MetaData: map[string]interface{}{
			"designation": "Saints Club Member",
		},
	}

	merchant := models.Merchant{
		BusinessName: "Saints Club",
		Phone:        "012XXXXXXXX",
		Email:        "contact@saintsclub.com",
		Settings: &models.MerchantSetting{
			BrandColor: "#BC925C",
			Address:    "QCH7+4GC, Road No.7, Dhaka 1212,",
		},
	}

	dgepayCard := dto.GenericResponse{
		Data: map[string]interface{}{
			"user_emp_cards_data": map[string]interface{}{
				"card_number": "1234 5678 9012 3456",
				"qr_code":     "dummy-saints-payment-qr",
				"merchant_info": map[string]interface{}{
					"business_name": "Saints Club",
				},
			},
		},
	}

	generateVDS := dto.GenerateVDSData{
		ExtractedQRImage: "",
	}

	images, err := BuildUniversityCardImages(emp, mem, merchant, dgepayCard, generateVDS)
	if err != nil {
		t.Fatalf("Failed to build saints club card images: %v", err)
	}

	if len(images.Thumbnail) == 0 {
		t.Fatal("Expected non-empty Thumbnail bytes")
	}

	thumbImg, _, err := image.Decode(bytes.NewReader(images.Thumbnail))
	if err != nil {
		t.Fatalf("Failed to decode generated thumbnail image: %v", err)
	}

	bounds := thumbImg.Bounds()
	if bounds.Dx() != saintsThumbnailW || bounds.Dy() != saintsThumbnailH {
		t.Fatalf("Expected thumbnail dimensions %dx%d, got %dx%d", saintsThumbnailW, saintsThumbnailH, bounds.Dx(), bounds.Dy())
	}

	if err := os.WriteFile("test_saints_thumbnail.png", images.Thumbnail, 0644); err != nil {
		t.Fatalf("Failed to write saints thumbnail: %v", err)
	}

	// Test 2: Two-line name thumbnail
	empLong := emp
	empLong.UserName = "Al Jabir Newton"
	imagesLong, err := BuildUniversityCardImages(empLong, mem, merchant, dgepayCard, generateVDS)
	if err != nil {
		t.Fatalf("Failed to build saints club card images with long name: %v", err)
	}
	if len(imagesLong.Thumbnail) == 0 {
		t.Fatal("Expected non-empty Thumbnail bytes for long name")
	}
	if err := os.WriteFile("test_saints_thumbnail_long.png", imagesLong.Thumbnail, 0644); err != nil {
		t.Fatalf("Failed to write saints thumbnail long name: %v", err)
	}

	t.Log("Successfully generated test_saints_thumbnail.png and test_saints_thumbnail_long.png")
}
