package utils

import (
	"io"
	"net/http"
	"os"
	"testing"
)

func fetchImageBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func TestMergeImagesWithRealURLs(t *testing.T) {
	oneIdURL := "https://res.cloudinary.com/dz1rfos2a/image/upload/v1778354547/1e5402cc-a31f-491b-8ac5-2b10433aebbf/1778354544-one-id.png"
	paymentURL := "https://res.cloudinary.com/dz1rfos2a/image/upload/v1778354547/1e5402cc-a31f-491b-8ac5-2b10433aebbf/1778354547-dgepay.png"

	t.Log("Downloading OneID image...")
	img1Bytes, err := fetchImageBytes(oneIdURL)
	if err != nil {
		t.Fatalf("Failed to fetch image 1: %v", err)
	}

	t.Log("Downloading DGEPay payment image...")
	img2Bytes, err := fetchImageBytes(paymentURL)
	if err != nil {
		t.Fatalf("Failed to fetch image 2: %v", err)
	}

	t.Log("Merging images...")
	mergedBytes, err := MergeImages(img1Bytes, img2Bytes)
	if err != nil {
		t.Fatalf("MergeImages returned an error: %v", err)
	}

	if len(mergedBytes) == 0 {
		t.Fatalf("Merged image bytes is empty")
	}

	outputPath := "merged_real_test_output.png"
	if err := os.WriteFile(outputPath, mergedBytes, 0644); err != nil {
		t.Fatalf("Failed to write merged output file: %v", err)
	}

	t.Logf("Success! Wrote merged image to %s", outputPath)
}

func TestMergeImagesWithHighResolutionCards(t *testing.T) {
	frontPath := "../services/test_student_front.png"
	backPath := "../services/test_student_back.png"

	img1Bytes, err := os.ReadFile(frontPath)
	if err != nil {
		t.Skip("High res test cards not found, skipping")
	}
	img2Bytes, err := os.ReadFile(backPath)
	if err != nil {
		t.Skip("High res test cards not found, skipping")
	}

	mergedBytes, err := MergeImages(img1Bytes, img2Bytes)
	if err != nil {
		t.Fatalf("MergeImages returned an error: %v", err)
	}

	maxAllowedBytes := 4194304 // 4MB
	if len(mergedBytes) > maxAllowedBytes {
		t.Fatalf("Merged image size (%d bytes) exceeds 4MB limit (%d bytes)", len(mergedBytes), maxAllowedBytes)
	}

	t.Logf("High resolution merge success! Output size: %d bytes (%.2f MB)", len(mergedBytes), float64(len(mergedBytes))/(1024*1024))
}
