package utils_test

import (
	"employee-card-admin/utils"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

const testIdentityToken = "eyJhbGciOiJSU0EtT0FFUC0yNTYiLCJjdHkiOiJhcHBsaWNhdGlvbi9qc29uIiwiZW5jIjoiQTI1NkdDTSIsImtpZCI6IlY2ODJFR3FNTmx2VUZ4dlJwSmY2dm1mNUxwTGtkLUN0OVk2VXBZU1ZpN00ifQ.eVJRQyYkKF3bBmwTu_ZRaYnbmeJmFbNH-AIg99ubFistdBMvCxzQOUtjqUUk0jFHhvEGcJDNQKH7M33zyoTPAExt1J_KBXUoWoVdtuFHr1E-zQS8x0lTDhqmJUAOLHO1EkbLaChXcSW_dJHoosACHyrti8MP7_R2KKHRh4xPlL3rXCzktsjOP9brwAGKeJT_ih6Bz20xhw8M-NLcgP_QTVo7B2_moIxtqQAbytqyt7HM8XXsjxw8OAA5HwtaQ0ngfn78rve6GGfYdAPSGX2j72Lhztayv-e60d7nieF9WNPPPXHbOUG57nrJQhZujm5WqUNImNpO1KQneXvUqNqiQA.SIVmKy0XZMXeAotI.HYWLoulEXvldTOhMWeWKvjm_whW0-CetOIqU_4Q5ex5vSWAIJceZ_Sqw.kfcYRHHDlJ2xY-g-SmaThQ"

func TestDecryptJWEIdentity(t *testing.T) {
	_ = godotenv.Load("../.env", "../.env.dev")
	key := os.Getenv("IDENTITY_PRIVATE_KEY")
	key = strings.Trim(strings.TrimSpace(key), `"`)
	key = strings.ReplaceAll(key, `\n`, "\n")
	if key == "" {
		t.Skip("IDENTITY_PRIVATE_KEY is not set in environment or .env, skipping test")
	}

	decrypted, err := utils.DecryptJWEIdentity(testIdentityToken, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if decrypted.NID != "7815160697" {
		t.Errorf("expected NID 7815160697, got %s", decrypted.NID)
	}

	if decrypted.DOB != "2000-03-08" {
		t.Errorf("expected DOB 2000-03-08, got %s", decrypted.DOB)
	}

	ts, err := utils.ConvertDOBToTimestamp(decrypted.DOB)
	if err != nil {
		t.Fatalf("failed to convert DOB to timestamp: %v", err)
	}
	if ts <= 0 {
		t.Errorf("invalid timestamp: %d", ts)
	}
}
