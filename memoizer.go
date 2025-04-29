package memoizer

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

// Memoizer is a structure that provides memoization capabilities.
// It stores results of expensive function calls and returns the cached result when possible.
type Memoizer[T any] struct {
	singleFlightGroup singleflight.Group
	cache             *cache.Cache
}

type unwrappableErr interface {
	Unwrap() error
}

// NewMemoizer creates and returns a new instance of a Memoizer.
func NewMemoizer[T any]() *Memoizer[T] {
	return &Memoizer[T]{
		singleFlightGroup: singleflight.Group{},
		cache:             cache.New(cache.NoExpiration, 0), // Initializes the cache with no expiration.
	}
}

// NewMemoizerWithCacheExpiration creates and returns a new instance of a Memoizer with a specified cache expiration time.
func NewMemoizerWithCacheExpiration[T any](expiration time.Duration) *Memoizer[T] {
	return &Memoizer[T]{
		singleFlightGroup: singleflight.Group{},
		cache:             cache.New(expiration, 0), // Initializes the cache with the specified expiration.
	}
}

type Options struct {
	Expiration    time.Duration
	CleanInterval time.Duration
}

// NewMemoizerWithOptions creates and returns a new instance of a Memoizer, with a specified cache expiration time and clean interval.
func NewMemoizerWithOptions[T any](opt Options) *Memoizer[T] {
	return &Memoizer[T]{
		singleFlightGroup: singleflight.Group{},
		cache:             cache.New(opt.Expiration, opt.CleanInterval), // Initializes the cache with the specified expiration.
	}
}

// Memoize checks the cache for a stored result for the given key. If not found, it executes the function,
// caches its result, and returns it. This method ensures that concurrent calls with the same key
// do not result in multiple executions of the function.
func (m *Memoizer[T]) Memoize(key string, fn func() (T, error), options ...Option) (T, error) {
	// Attempt to retrieve the cached value.
	value, ok := m.cache.Get(key)
	if ok {
		// If a value is found, assert its type and return it.
		typedValue, ok := value.(T)
		if !ok {
			panic(fmt.Errorf("cache value type mismatch"))
		}
		return typedValue, nil
	}

	defer func() {
		if r := recover(); r != nil {
			if ue, ok := r.(unwrappableErr); ok {
				panic(ue.Unwrap())
			} else {
				panic(r)
			}
		}
	}()

	// If no cached value is found, use singleflight to call the function and store its result.
	result, err, _ := m.singleFlightGroup.Do(key, func() (interface{}, error) {
		res, err := fn()
		if err == nil {
			// Cache the result if there's no error.
			expiration := cache.DefaultExpiration
			for _, option := range options {
				if opt, ok := option.(*ExpirationOption); ok {
					expiration = opt.Callback(res)
				}
			}
			m.cache.Set(key, res, expiration)
		}
		return res, err
	})

	if err != nil && result == nil {
		var zero T
		return zero, err
	}

	return result.(T), err
}
