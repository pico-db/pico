package breaker

import (
	"fmt"
	"sync"
	"time"
)

type State int8

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

// Implements
//
//	stringer
//
// interface
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateHalfOpen:
		return "half-open"
	case StateOpen:
		return "open"
	default:
		return fmt.Sprintf("invalid state: %d", s)
	}
}

var (
	ErrCircuitOpen     = NewBreakerError("circuit open")
	ErrTooManyRequests = NewBreakerError("number of requests exceeded the circuit limit")
)

var (
	ReasonStateChange = NewResetReason("state changes")
	ReasonClosedReset = NewResetReason("reset counter in Closed state")
)

type StateChangeCallback func(name string, prev State, now State)
type IsSuccessCallback func(err error) bool
type Breaker func(stats Statistics) bool
type ResetCallback func(stats Statistics, reason ResetReason)

type Options struct {
	// The name of this Circuit Breaker
	Name string `json:"name"`

	// The maximum number of task allowed to be performed during Half-open state
	MaxDiscoveryRequests uint32

	// The timeout of the Open state.
	// After which, the state transitions to Half-open
	OpenTimeout time.Duration

	// The interval for resetting count during Closed state.
	// For example, the transition to Open should only occur after failed attempts reaches 5, in a 10 seconds window.
	// To achieve that, you can set this to 10 seconds and set ShouldBreakCircuit to returns a breaker that breaks the circuit after 5 failures.
	//
	// By default, the interval is 0, meaning that it doesn't reset the count
	ResetCountInterval time.Duration

	// The callback to be called whenever the internal state changes
	OnStateChange StateChangeCallback

	// The callback to be called on counter reset
	OnCountReset ResetCallback

	// This determines if a task is success or not, based on the returned error.
	// By default, all errors returned from the task indicate a failed task.
	IsSuccess IsSuccessCallback

	// This callback is used to determine when a Closed circuit should transitions to Open state.
	// By default, it should breaks the circuit when the number of failed attempts reaches 5.
	// Runs only when the task returns an error
	ShouldBreakCircuit Breaker
}

type Statistics struct {
	Requests  uint32
	Successes uint32
	Failures  uint32

	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

// A state machine  to prevent sending requests that are likely to fails
type CircuitBreaker struct {
	name        string
	maxRequests uint32
	interval    time.Duration
	timeout     time.Duration

	sync.Mutex
	state            State
	openExpiry       time.Time
	closedResetTimer time.Time
	count            Statistics

	// Optional callbacks
	isSuccessful  IsSuccessCallback
	onStateChange StateChangeCallback
	breaker       Breaker
	onResetCount  ResetCallback
}

func (s *Statistics) onSuccess() {
	s.Successes += 1
	s.ConsecutiveSuccesses += 1
	s.ConsecutiveFailures = 0
}

func (s *Statistics) onFailure() {
	s.Failures += 1
	s.ConsecutiveFailures += 1
	s.ConsecutiveSuccesses = 0
}

func (s *Statistics) reset() {
	s.Successes = 0
	s.ConsecutiveFailures = 0
	s.ConsecutiveSuccesses = 0
	s.Failures = 0
	s.Requests = 0
}

func (s *Statistics) onRequest() {
	s.Requests += 1
}

// Provided an Options, it returns the new Circuit Breaker
func New(opts Options) *CircuitBreaker {
	c := &CircuitBreaker{}
	c.name = opts.Name
	c.onStateChange = opts.OnStateChange
	c.isSuccessful = opts.IsSuccess
	c.count = Statistics{}
	c.onResetCount = opts.OnCountReset

	if opts.MaxDiscoveryRequests > 0 {
		c.maxRequests = opts.MaxDiscoveryRequests
	} else {
		c.maxRequests = 1
	}

	if opts.OpenTimeout > 0 {
		c.timeout = opts.OpenTimeout
	} else {
		c.timeout = time.Second * 60
	}

	if opts.ResetCountInterval > 0 {
		c.interval = opts.ResetCountInterval
	} else {
		c.interval = time.Second * 0
	}

	if opts.ShouldBreakCircuit != nil {
		c.breaker = opts.ShouldBreakCircuit
	} else {
		c.breaker = func(stats Statistics) bool {
			return stats.ConsecutiveFailures >= 5
		}
	}

	return c
}

// Perform the action if the Circuit Breaker allows it, either in Closed state of in Half-open if the number requests does not exceed the limit.
// It returns an error of type BreakerError if the Circuit Breaker rejects the task.
// It treats any panic as regular error and forwards it back to the consumer.
func (c *CircuitBreaker) Do(what func() (interface{}, error)) (interface{}, error) {
	err := c.onTaskStarted()
	if err != nil {
		return nil, err
	}
	defer func() {
		pan := recover()
		if pan != nil {
			c.onTaskFinished(false)
			panic(pan)
		}
	}()
	res, err := what()
	if c.isSuccessful != nil {
		success := true
		if err != nil {
			success = false
		}
		c.onTaskFinished(success)
	}
	return res, err
}

func (c *CircuitBreaker) onTaskStarted() error {
	c.Lock()
	defer c.Unlock()
	now := time.Now()
	state, _ := c.getState(now)

	switch state {
	case StateOpen:
		// If the circuit was open or is now open
		return ErrCircuitOpen
	case StateHalfOpen:
		if c.count.Requests >= c.maxRequests {
			// In Half-open state, only a certain number of requests are permitted
			return ErrTooManyRequests
		}
	default: // StateClosed
		c.count.onRequest()
	}
	return nil
}

func (c *CircuitBreaker) onTaskFinished(isSuccessful bool) {
	c.Lock()
	defer c.Unlock()

	now := time.Now()
	_, didResetCounter := c.getState(now)
	if didResetCounter {
		// Prevent the end of the current counter cycle
		// being also the start of the next counter cycle
		return
	}

	if isSuccessful {
		c.onSuccess()
	} else {
		c.onFailure()
	}
}

// In case of a successful task, transitions Half-open to Close
// when the number of consecutive success exceeds the threshold
func (c *CircuitBreaker) onSuccess() {
	switch c.state {
	case StateClosed:
		c.count.onSuccess()
	case StateHalfOpen:
		c.count.onSuccess()
		if c.count.ConsecutiveSuccesses >= c.maxRequests {
			c.setState(StateClosed)
		}
	default: // StateOpen
	}
}

// In case a request fails, transitions all to Open state
func (c *CircuitBreaker) onFailure() {
	switch c.state {
	case StateClosed:
		c.count.onFailure()
		// Determines if the circuit needs to be broken after a failure
		if c.breaker(c.count) {
			c.setState(StateOpen)
		}
	case StateHalfOpen:
		c.setState(StateOpen)
	default: // StateOpen
	}
}

func (c *CircuitBreaker) getState(now time.Time) (State, bool) {
	switch c.state {
	case StateClosed:
		// Checks if the Closed state timer is up.
		// If so, reset the counter and reset the timer
		if !c.closedResetTimer.IsZero() && c.closedResetTimer.Before(now) {
			c.reset()
			if c.onResetCount != nil {
				c.onResetCount(c.count, ReasonClosedReset)
			}
			return c.state, true
		}
	case StateOpen:
		// Checks if the Circuit Breaker passes the Open timeout duration.
		// If yes, move it to Half-open state
		if c.openExpiry.Before(now) {
			c.setState(StateHalfOpen)
		}
	}
	return c.state, false
}

// Update the internal state of the Circuit Breaker
// and executes the optional onStateChange function
func (c *CircuitBreaker) setState(s State) {
	if c.state == s {
		return
	}

	prev := c.state
	c.state = s

	// On new state
	if c.onResetCount != nil {
		c.onResetCount(c.count, ReasonStateChange)
	}
	c.reset()

	if c.onStateChange != nil {
		c.onStateChange(c.name, prev, c.state)
	}
}

func (c *CircuitBreaker) reset() {
	c.count.reset()
	switch c.state {
	case StateOpen:
		// Resets the timer if the new CB is now Open:
		c.openExpiry = time.Now().Add(c.timeout)
	case StateClosed:
		// Add a new timer for Closed state count reset
		if c.interval == 0 {
			c.closedResetTimer = time.Time{}
		} else {
			c.closedResetTimer = time.Now().Add(c.interval)
		}
	default: // StateHalfOpen
		// Does not matter. Resets timer to zero to prevent wasteful clock cycles
		c.openExpiry = time.Time{}
		c.closedResetTimer = time.Time{}
	}
}
