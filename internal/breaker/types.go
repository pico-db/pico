package breaker

// Contains the reason for a reset in the breaker's internal counter
type ResetReason struct {
	reason string
}

func NewResetReason(r string) ResetReason {
	return ResetReason{
		reason: r,
	}
}

// Returns the reason of the reset
func (r *ResetReason) String() string {
	return r.reason
}
