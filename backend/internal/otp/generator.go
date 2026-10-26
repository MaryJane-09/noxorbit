package otp

import (
	"crypto/rand"
	"math/big"
)

func Generate(length int) (string, error) {

	const charset = "abcdefghjklmnpqrstuvwxyz23456789"

	otp := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		randomIdx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		otp[i] = charset[int(randomIdx.Int64())]
	}
	return string(otp), nil
}
