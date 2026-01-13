package service

import (
	"crypto/hmac"
	"crypto/sha256"
)

var base62 = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type Hasher struct {
	Secret string
}

func NewHasher(secret string) *Hasher {
	return &Hasher{Secret: secret}
}

func (h *Hasher) GetHash(input string) string {
	mac := hmac.New(sha256.New, []byte(h.Secret))
	mac.Write([]byte(input))
	sum := mac.Sum(nil) // 32 bytes

	// Use first 6 bytes (48 bits)
	v := uint64(sum[0])<<40 | uint64(sum[1])<<32 | uint64(sum[2])<<24 |
		uint64(sum[3])<<16 | uint64(sum[4])<<8 | uint64(sum[5])

	// Map 48-bit value into 62^8 space and encode Base62 (8 chars)
	const space = 218340105584896 // 62^8
	v = v % space

	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = base62[v%62]
		v /= 62
	}
	return string(buf)
}
