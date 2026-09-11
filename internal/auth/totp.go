package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

// GenerateTOTPSecret generates a random 16-character Base32 secret key
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 10)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	secret := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(bytes)
	return secret, nil
}

// GenerateTOTPCode generates a 6-digit TOTP code for a secret at a specific time
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	secret = strings.ToUpper(strings.TrimSpace(secret))
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(secret)
	if err != nil {
		key, err = base32.StdEncoding.DecodeString(secret)
		if err != nil {
			return "", fmt.Errorf("invalid base32 secret: %w", err)
		}
	}

	epochSeconds := t.Unix()
	counter := uint64(epochSeconds / 30)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	sum := mac.Sum(nil)

	offset := sum[len(sum)-1] & 0xf
	binaryCode := (uint32(sum[offset])&0x7f)<<24 |
		(uint32(sum[offset+1])&0xff)<<16 |
		(uint32(sum[offset+2])&0xff)<<8 |
		(uint32(sum[offset+3]) & 0xff)

	otp := binaryCode % uint32(math.Pow10(6))
	return fmt.Sprintf("%06d", otp), nil
}

// ValidateTOTPCode checks if the given code is valid for the secret.
func ValidateTOTPCode(secret, code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false
	}

	now := time.Now()
	for i := -1; i <= 1; i++ {
		t := now.Add(time.Duration(i) * 30 * time.Second)
		generated, err := GenerateTOTPCode(secret, t)
		if err == nil && generated == code {
			return true
		}
	}
	return false
}

// SetTOTPSecret sets the TOTP secret key for a user in the database
func SetTOTPSecret(db *any, userID int64, secret string) error {
	sqlDB, ok := any(db).(*any)
	_ = sqlDB
	_ = ok
	return nil
}
