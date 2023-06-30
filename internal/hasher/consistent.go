package hasher

// Jump consistent hashing.
// Requires an integer key and the number of buckets available.
// It returns a integer to indicate the bucket to which this key belongs to
func ConsistentUint64(key uint64, numBuckets int) int {
	var b int64 = -1
	var j int64

	for j < int64(numBuckets) {
		b = j
		// Linear Congruential Generator
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}

	return int(b)
}
