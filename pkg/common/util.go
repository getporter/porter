package common

import (
	"crypto/sha256"
	"encoding/binary"
	"math/rand"
)

// RandomString generates a hash string with fixed length from "a-zA-Z0-9" using
// a seed string and provided length.
func RandomString(seed string, length int) string {
	h := sha256.New()
	h.Write([]byte(seed))
	rand.Seed(int64(binary.BigEndian.Uint64(h.Sum(nil))))
	b := make([]byte, length)
	charset := "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	for i := 0; i < length; i++ {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
