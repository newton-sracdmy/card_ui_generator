package utils

import (
	"context"
	"employee-card-admin/models"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

func DecodeImageBytes(ctx context.Context, payload string) ([]byte, string, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return nil, "", errors.New("empty image payload")
	}

	if strings.HasPrefix(payload, "http://") || strings.HasPrefix(payload, "https://") {
		data, err := DownloadFile(ctx, http.DefaultClient, payload)
		if err != nil {
			return nil, "", err
		}
		ext := strings.TrimPrefix(filepath.Ext(payload), ".")
		if ext == "" {
			ext = "png"
		}
		return data, ext, nil
	}

	rawB64 := payload
	ext := "png"

	if strings.HasPrefix(payload, "data:") {
		parts := strings.SplitN(payload, ",", 2)
		if len(parts) != 2 {
			return nil, "", errors.New("invalid data URI format")
		}
		ext = DetectImageExtension(parts[0])
		rawB64 = parts[1]
	}

	cleanB64 := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == ' ' {
			return -1
		}
		return r
	}, rawB64)

	data, err := base64.StdEncoding.DecodeString(cleanB64)
	if err != nil {
		data, err = base64.RawStdEncoding.DecodeString(cleanB64)
		if err != nil {
			return nil, "", fmt.Errorf("invalid base64 encoding: %w", err)
		}
	}

	if len(data) == 0 {
		return nil, "", errors.New("decoded image data is empty")
	}

	return data, ext, nil
}

func DetectImageExtension(header string) string {
	header = strings.ToLower(header)
	switch {
	case strings.Contains(header, "image/jpeg"), strings.Contains(header, "image/jpg"):
		return "jpg"
	case strings.Contains(header, "image/webp"):
		return "webp"
	case strings.Contains(header, "image/gif"):
		return "gif"
	default:
		return "png"
	}
}

func SanitizeFilename(name string) string {
	clean := strings.TrimSpace(name)
	if clean == "" {
		return fmt.Sprintf("file_%d", time.Now().Unix())
	}
	replacer := strings.NewReplacer(" ", "_", "/", "_", "\\", "_")
	return replacer.Replace(clean)
}

func ExtractAvatarPayload(meta models.JSONMap, fallbackURL string) string {
	formData := GetFormDataFromMeta(meta)
	if formData != nil {
		switch val := formData["upload_your_picture"].(type) {
		case string:
			if str := strings.TrimSpace(val); str != "" {
				return str
			}
		case map[string]interface{}:
			for _, key := range []string{"content", "value", "url"} {
				if str, ok := val[key].(string); ok && strings.TrimSpace(str) != "" {
					return strings.TrimSpace(str)
				}
			}
		}
	}
	return strings.TrimSpace(fallbackURL)
}

func SanitizeAvatarInMeta(meta models.JSONMap, publicID string) {
	formData := GetFormDataFromMeta(meta)
	if formData == nil {
		return
	}
	switch val := formData["upload_your_picture"].(type) {
	case map[string]interface{}:
		val["public_id"] = publicID
		val["content"] = publicID
		val["value"] = publicID
	default:
		formData["upload_your_picture"] = publicID
	}
}

func GetFormDataFromMeta(meta models.JSONMap) map[string]interface{} {
	if meta == nil {
		return nil
	}
	for _, key := range []string{"form_data", "formdata", "formData"} {
		if fd, ok := meta[key].(map[string]interface{}); ok {
			return fd
		}
	}
	return nil
}

func LookupStringFromMeta(meta models.JSONMap, formData map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if val := LookupStringInMap(map[string]interface{}(meta), key); val != "" {
			return val
		}
		if val := LookupStringInMap(formData, key); val != "" {
			return val
		}
	}
	return ""
}

func LookupStringInMap(m map[string]interface{}, targetKey string) string {
	if m == nil {
		return ""
	}

	if val, ok := m[targetKey]; ok {
		if s := ParseStringValue(val); s != "" {
			return s
		}
	}
	for k, val := range m {
		if strings.EqualFold(k, targetKey) {
			if s := ParseStringValue(val); s != "" {
				return s
			}
		}
	}
	return ""
}

func ParseStringValue(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return strings.TrimSpace(v)
	case map[string]interface{}:
		if str, ok := v["value"].(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return ""
}
