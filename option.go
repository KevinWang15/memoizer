package memoizer

import "time"

// Option is an interface for configuring the Memoizer's behavior.
// Implementations of this interface can be passed to the Memoize function
// to customize its operation.
type Option interface {
}

// ExpirationOption is a struct that implements the Option interface.
// It contains a Callback function that determines the expiration duration
// for a cached result based on the result itself.
type ExpirationOption struct {
	Callback func(result interface{}) time.Duration
}

// WithExpiration returns an Option that sets a dynamic expiration time for cached results.
// The provided callback function is called with the result of the memoized function
// and should return a time.Duration indicating how long the result should be cached.
//
// Example usage:
//
//	memoizer.Memoize("key", myFunc, memoizer.WithExpiration(func(result interface{}) time.Duration {
//	    // Custom logic to determine expiration based on the result
//	    return time.Hour
//	}))
var WithExpiration = func(callback func(result interface{}) time.Duration) Option {
	return &ExpirationOption{Callback: callback}
}
