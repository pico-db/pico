package utils

import (
	"math/rand"
	"time"
)

// Thanks icza and moorana from StackOverflow
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var randSrc rand.Source

// init is guaranteed to be called once
func init() {
	randSrc = rand.NewSource(time.Now().UnixNano())
}

// Generate a random string
func RandomString(length int) string {
	b := randomBytes(length)
	return string(b)
}

// I honestly don't understand this marvelous beast
// But it's fast
func randomBytes(length int) []byte {
	b := make([]byte, length)
	for i, cache, rem := (length - 1), randSrc.Int63(), letterIdxMax; i >= 0; {
		if rem == 0 {
			cache, rem = randSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i -= 1
		}
		cache >>= letterIdxBits
		rem -= 1
	}
	return b
}
