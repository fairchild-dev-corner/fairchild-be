package services

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const (
	otpCodeLength  = 6
	otpExpiry      = 5 * time.Minute
	otpMaxAttempts = 5
)

// GenerateOTPCode produces a zero-padded 6-digit numeric code using a
// uniform random draw (crypto/rand.Int, not rand.Read+modulo, which would
// introduce modulo bias).
func GenerateOTPCode() (string, error) {
	max := big.NewInt(1000000) // 10^otpCodeLength
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
