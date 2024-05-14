package memoizer

import (
	"fmt"

	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

// Memoizer is a structure that provides memoization capabilities.
// It stores results of expensive function calls and returns the cached result when possible.
type Memoizer[T any] struct {
	singleFlightGroup singleflight.Group
	cache             *cache.Cache
}

// NewMemoizer creates and returns a new instance of a Memoizer.
func NewMemoizer[T any]() *Memoizer[T] {
	return &Memoizer[T]{
		singleFlightGroup: singleflight.Group{},
		cache:             cache.New(cache.NoExpiration, 0), // Initializes the cache with no expiration.
	}
}

// Memoize checks the cache for a stored result for the given key. If not found, it executes the function,
// caches its result, and returns it. This method ensures that concurrent calls with the same key
// do not result in multiple executions of the function.
func (m *Memoizer[T]) Memoize(key string, fn func() (T, error)) (T, error) {
	var zero T // Zero value for type T, used as a default return value.

	// Attempt to retrieve the cached value.
	value, ok := m.cache.Get(key)
	if ok {
		// If a value is found, assert its type and return it.
		typedValue, ok := value.(T)
		if !ok {
			return zero, fmt.Errorf("cache value type mismatch")
		}
		return typedValue, nil
	}

	// If no cached value is found, use singleflight to call the function and store its result.
	result, err, _ := m.singleFlightGroup.Do(key, func() (interface{}, error) {
		res, err := fn()
		if err == nil {
			m.cache.Set(key, res, cache.DefaultExpiration) // Cache the result if there's no error.
		}
		return res, err
	})

	if err != nil {
		return zero, err // Return error if the function execution failed.
	}

	// Assert the type of the result before returning it.
	typedResult, ok := result.(T)
	if !ok {
		return zero, fmt.Errorf("result type mismatch")
	}

	return typedResult, nil
}
