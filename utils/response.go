package utils

import (
	"fmt"
	"strings"

	"time"

	"github.com/go-playground/validator/v10"
)

// UTC+6 for Bangladesh Standard Time — no external tzdata dependency
var bdTime = time.FixedZone("Asia/Dhaka", 6*60*60)

type Response[T any] struct {
	Status     bool   `json:"status"`
	Message    string `json:"message"`
	Data       T      `json:"data"`
	Error      any    `json:"error,omitempty"`
	ServerTime string `json:"server_time"`
	LocalTime  string `json:"local_time"`
}

func SuccessResponse[T any](data T, message string) Response[T] {
	return Response[T]{
		Status:     true,
		Message:    message,
		Data:       data,
		Error:      nil,
		ServerTime: time.Now().Format("2006-01-02 15:04:05"),
		LocalTime:  time.Now().In(bdTime).Format("2006-01-02 15:04:05"),
	}
}

func ErrorResponse(message string, err any) Response[any] {
	fmt.Println("ErrorResponse: ", message, err)

	var errorData any

	// Handle validator.ValidationErrors
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errorMap := make(map[string]string)
		for _, fieldError := range validationErrors {
			errorMap[strings.ToLower((fieldError.Field()))] = fieldError.Tag()
		}
		errorData = errorMap
	} else {
		// For other error types, try to extract message
		switch v := err.(type) {
		case string:
			errorData = map[string]string{"message": v}
		case error:
			errorData = map[string]string{"message": v.Error()}
		case map[string]any:
			errorData = v
		default:
			errorData = map[string]string{"message": fmt.Sprintf("%v", v)}
		}
	}

	return Response[any]{
		Status:     false,
		Message:    message,
		Error:      errorData,
		ServerTime: time.Now().Format("2006-01-02 15:04:05"),
		LocalTime:  time.Now().In(bdTime).Format("2006-01-02 15:04:05"),
	}
}
