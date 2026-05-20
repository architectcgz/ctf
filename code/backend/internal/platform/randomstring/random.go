package randomstring

import (
	"crypto/rand"
	"encoding/base64"
)

const DefaultEntropyBytes = 32

func Generate() (string, error) {
	return GenerateWithEntropyBytes(DefaultEntropyBytes)
}

func GenerateWithEntropyBytes(entropyBytes int) (string, error) {
	bytes := make([]byte, entropyBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
