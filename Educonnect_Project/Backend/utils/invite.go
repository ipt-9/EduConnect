package utils

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

func GenerateInviteCode() string {
	b := make([]byte, 5)
	_, _ = rand.Read(b)
	return strings.ToUpper(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b))
}
