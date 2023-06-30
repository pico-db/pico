package breaker

// Implements the error interface
type BreakerError struct {
	message string
}

func NewBreakerError(m string) *BreakerError {
	return &BreakerError{
		message: m,
	}
}

func (e BreakerError) Error() string {
	return e.message
}
