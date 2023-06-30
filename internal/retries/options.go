package retries

import (
	"context"
	"math"
	"time"
)

type Option func(*Config)
type OnRetryCallback func(trial uint, incomingTimeout time.Duration, err error)
type RetryDelegate func(err error) bool
type DelayType func(n uint, err error, c *Config) time.Duration

var (
	emptyOption Option = func(c *Config) {}
)

type Config struct {
	attempts   uint
	maxBackoff uint
	delay      time.Duration
	maxDelay   time.Duration
	onRetry    OnRetryCallback
	retryIf    RetryDelegate
	ctx        context.Context
	delayType  DelayType
}

// Set a hard concrete number of retries.
// Default is 3. To retry until working, set to a number less than 1
func Attempts(number uint) Option {
	return func(c *Config) {
		c.attempts = number
	}
}

// Set a delay between retries.
// Default is 100ms
func Delay(delay time.Duration) Option {
	if delay <= 0 {
		delay = time.Millisecond * 100
	}
	return func(c *Config) {
		c.delay = delay
	}
}

// Set the maximum value for a retry.
func MaxDelay(max time.Duration) Option {
	if max < 0 {
		max = 0
	}
	return func(c *Config) {
		c.maxDelay = max
	}
}

// Set a callback to be called on retries.
func OnRetry(callback OnRetryCallback) Option {
	if callback == nil {
		return emptyOption
	}
	return func(c *Config) {
		c.onRetry = callback
	}
}

// Set a delegate that determines when to retry
func RetryIf(del RetryDelegate) Option {
	if del == nil {
		return func(c *Config) {
			c.retryIf = func(err error) bool {
				return true
			}
		}
	}
	return func(c *Config) {
		c.retryIf = del
	}
}

// Set a context for the execution.
// If it has a timeout, then it will cancel the execution when the timeout is up
func Context(ctx context.Context) Option {
	return func(c *Config) {
		c.ctx = ctx
	}
}

// Set the type of delay to use.
// Default is ConstantDelay
func DelayMethod(typ DelayType) Option {
	if typ == nil {
		return func(c *Config) {
			c.delayType = ConstantDelay
		}
	}
	return func(c *Config) {
		c.delayType = typ
	}
}

// The type of delay that have a constant timeout between retries.
// The delay is specified through Delay option
func ConstantDelay(n uint, err error, c *Config) time.Duration {
	return c.delay
}

// This type of delay doubles the timeout between retries as long as it does not overflow int64 of time.Duration
func BackoffDelay(n uint, err error, c *Config) time.Duration {
	delay := c.delay
	if delay <= 0 {
		delay = time.Millisecond * 100
	}
	// double each times
	// can cause overflow as bits shifted to the left
	// need to calculate the minimum number of binary digits to represent the current number
	if c.maxBackoff == 0 {
		maxDigit := uint(math.Ceil(math.Log2(float64(c.delay))))
		c.maxBackoff = 63 - maxDigit
	}
	if n > c.maxBackoff {
		n = c.maxBackoff
	}
	return c.delay << n
}
