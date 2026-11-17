package utils

import (
	"crypto/rand"
	"employee-card-admin/common/constant"
	"math/big"
	"strings"
)

func IsYogiDesignation(designation string) bool {
	return strings.EqualFold(strings.TrimSpace(designation), constant.DesignationYogi)
}


func SplitFullName(fullName string) (string, string) {
	fullName = strings.TrimSpace(fullName)
	if idx := strings.LastIndex(fullName, " "); idx != -1 {
		return strings.TrimSpace(fullName[:idx]), strings.TrimSpace(fullName[idx+1:])
	}
	return fullName, ""
}


func ToBDInternationalPhone(phone string) string {
	phone = strings.TrimSpace(phone)
	switch {
	case strings.HasPrefix(phone, "+"):
		return phone
	case strings.HasPrefix(phone, "880"):
		return "+" + phone
	case strings.HasPrefix(phone, "0"):
		return "+88" + phone
	default:
		return "+880" + phone
	}
}


func GenerateSecurePassword(length int) (string, error) {
	const (
		upper   = "ABCDEFGHJKLMNPQRSTUVWXYZ"
		lower   = "abcdefghijkmnpqrstuvwxyz"
		digits  = "23456789"
		symbols = "!@#$%^&*"
	)
	sets := []string{upper, lower, digits, symbols}
	all := upper + lower + digits + symbols
	if length < len(sets) {
		length = len(sets)
	}

	pick := func(charset string) (byte, error) {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return 0, err
		}
		return charset[n.Int64()], nil
	}

	password := make([]byte, length)
	for i := range password {
		charset := all
		if i < len(sets) {
			charset = sets[i]
		}
		c, err := pick(charset)
		if err != nil {
			return "", err
		}
		password[i] = c
	}

	for i := len(password) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return "", err
		}
		j := n.Int64()
		password[i], password[j] = password[j], password[i]
	}

	return string(password), nil
}
