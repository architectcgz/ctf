package commands

import (
	"crypto/rand"
	"encoding/hex"
)

func generateChallengeImportPreviewID() (string, error) {
	token := make([]byte, 12)
	if _, err := rand.Read(token); err != nil {
		return "", err
	}
	return hex.EncodeToString(token), nil
}
