package crypt

import (
	"crypto/sha256"
)

func GenX25519Private(data []byte) []byte {
	key := sha256.Sum256(data)
	key[0] &= 248
	key[31] &= 127
	key[31] |= 64
	return key[:32]
}
