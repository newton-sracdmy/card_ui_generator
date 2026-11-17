package card_generator

import (
	"bytes"
	"employee-card-admin/dto"
	"employee-card-admin/models"
	"image"
	"os"
	"testing"
	"time"
)

func TestResetStudioMemberCardUIGeneration(t *testing.T) {
	emp := models.Employee{
		ID:        "reset-studio-test-emp-1",
		UserName:  "Aarusha Zahia",
		OneID:     "1234567890",
		AvatarURL: "https://res.cloudinary.com/llhcc8la/image/upload/v1786441014/Card_UI/1786441013-newton-photo%20%281%29.jpg.jpg",
		Phone:     "+8801700000000",
		Email:     "aarusha@resetstudio.com",
	}

	mem := models.Membership{
		ID:          "reset-studio-test-mem-1",
		RefID:       "XXXX XXXX 1234",
		Designation: "Reset Studio Member",
		MetaData: models.JSONMap{
			"designation": "Reset Studio Member",
		},
	}

	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	merchant := models.Merchant{
		BusinessName: "Reset Studio",
		Phone:        "+880 1865-840108",
		Email:        "contact@resetstudio.com",
		Settings: &models.MerchantSetting{
			BrandColor:     "#8B8D68",
			Address:        "Road 73, House 4D, 5th floor, Gulshan 2, Dhaka-1212",
			CreatedAt:      now,
			CardExpiryTime: now.AddDate(1, 0, 0),
		},
	}

	dgepayCard := dto.GenericResponse{
		Data: map[string]interface{}{
			"user_emp_cards_data": map[string]interface{}{
				"card_number": "XXXX XXXX 1234",
				"qr_code":     "dummy-reset-studio-payment-qr",
				"merchant_info": map[string]interface{}{
					"business_name": "Reset Studio",
				},
			},
		},
	}

	generateVDS := dto.GenerateVDSData{
		ExtractedQRImage: "",
	}

	images, err := BuildResetStudioMemberCardImages(emp, mem, merchant, dgepayCard, generateVDS)
	if err != nil {
		t.Fatalf("Failed to build reset studio card images: %v", err)
	}

	if len(images.Payment) == 0 {
		t.Fatal("Expected non-empty Payment (front) card bytes")
	}
	if len(images.OneID) == 0 {
		t.Fatal("Expected non-empty OneID (back) card bytes")
	}
	if len(images.Thumbnail) == 0 {
		t.Fatal("Expected non-empty Thumbnail bytes")
	}

	frontImg, _, err := image.Decode(bytes.NewReader(images.Payment))
	if err != nil {
		t.Fatalf("Failed to decode reset studio front image: %v", err)
	}
	if b := frontImg.Bounds(); b.Dx() != resetCardW || b.Dy() != resetCardH {
		t.Fatalf("Expected front dimensions %dx%d, got %dx%d", resetCardW, resetCardH, b.Dx(), b.Dy())
	}

	thumbImg, _, err := image.Decode(bytes.NewReader(images.Thumbnail))
	if err != nil {
		t.Fatalf("Failed to decode reset studio thumbnail image: %v", err)
	}
	if b := thumbImg.Bounds(); b.Dx() != resetThumbnailW || b.Dy() != resetThumbnailH {
		t.Fatalf("Expected thumbnail dimensions %dx%d, got %dx%d", resetThumbnailW, resetThumbnailH, b.Dx(), b.Dy())
	}

	if err := os.WriteFile("test_reset_studio_front.png", images.Payment, 0644); err != nil {
		t.Fatalf("Failed to write reset studio front card: %v", err)
	}
	if err := os.WriteFile("test_reset_studio_back.png", images.OneID, 0644); err != nil {
		t.Fatalf("Failed to write reset studio back card: %v", err)
	}
	if err := os.WriteFile("test_reset_studio_thumbnail.png", images.Thumbnail, 0644); err != nil {
		t.Fatalf("Failed to write reset studio thumbnail: %v", err)
	}

	// Two-line name variation for thumbnail/front wrapping
	empLong := emp
	empLong.UserName = "Aarusha Zahia Chowdhury Khan"
	imagesLong, err := BuildResetStudioMemberCardImages(empLong, mem, merchant, dgepayCard, generateVDS)
	if err != nil {
		t.Fatalf("Failed to build reset studio card images with long name: %v", err)
	}
	if len(imagesLong.Thumbnail) == 0 {
		t.Fatal("Expected non-empty Thumbnail bytes for long name")
	}
	if err := os.WriteFile("test_reset_studio_thumbnail_long.png", imagesLong.Thumbnail, 0644); err != nil {
		t.Fatalf("Failed to write reset studio thumbnail long name: %v", err)
	}

	t.Log("Successfully generated test_reset_studio_front.png, test_reset_studio_back.png, test_reset_studio_thumbnail.png and test_reset_studio_thumbnail_long.png")
}
