package auth

import (
	"crypto/rand"
	"fmt"
	"time"
	"math/big"
)

func GenerateOTP() (string, time.Time, error) {
	max := big.NewInt(1000000) // 0..999999
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", time.Time{}, err
	}

	otp := fmt.Sprintf("%06d", n.Int64())
	expiry := time.Now().Add(10 * time.Minute)
	return otp, expiry, nil
}
