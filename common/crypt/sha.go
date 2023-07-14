package crypt

import (
	"crypto/sha256"
	"encoding/hex"
)

func GenShaHash(data []byte) string {
	d := sha256.Sum256(data)
	return hex.EncodeToString(d[:])
}
