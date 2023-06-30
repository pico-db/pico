package hasher

import (
	"github.com/spaolacci/murmur3"
)

// Use Murmur3 hash algorithm to hash bytes into an uint64
func MurmurToUint64(s []byte) uint64 {
	return murmur3.Sum64(s)
}
