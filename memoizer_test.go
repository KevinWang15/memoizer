package memoizer

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemoizer(t *testing.T) {
	t.Run("Basic Functionality", testBasicFunctionality)
	t.Run("Error Handling", testErrorHandling)
	t.Run("Concurrent Memoization", testConcurrentMemoization)
	t.Run("Panic Propagation", testPanicPropagation)
	t.Run("No Memoization on Error", testNoMemoizationOnError)
	t.Run("No Memoization on Panic", testNoMemoizationOnPanic)
}

func testBasicFunctionality(t *testing.T) {
	memoizer := NewMemoizer[int]()

	// First call
	result1, err := memoizer.Memoize("key", func() (int, error) {
		return 42, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 42, result1)

	// Second call (should return cached result)
	result2, err := memoizer.Memoize("key", func() (int, error) {
		return 0, errors.New("this should not be called")
	})
	require.NoError(t, err)
	assert.Equal(t, 42, result2)
}

func testErrorHandling(t *testing.T) {
	memoizer := NewMemoizer[int]()

	_, err := memoizer.Memoize("error_key", func() (int, error) {
		return 0, errors.New("intentional error")
	})
	require.Error(t, err)
	assert.EqualError(t, err, "intentional error")

	memoizer2 := NewMemoizer[any]()

	a, b := memoizer2.Memoize("A", func() (any, error) {
		return nil, fmt.Errorf("A")
	})
	require.Error(t, err)
	assert.Equal(t, a, nil)
	assert.EqualError(t, b, "A")
}

func testConcurrentMemoization(t *testing.T) {
	memoizer := NewMemoizer[int]()
	var wg sync.WaitGroup
	const goroutines = 10

	invocations := 0
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := memoizer.Memoize("concurrent_key", func() (int, error) {
				invocations++
				time.Sleep(10 * time.Millisecond) // Simulate work
				return 100, nil
			})
			require.NoError(t, err)
			assert.Equal(t, 100, result)
		}()
	}

	wg.Wait()
	assert.Equal(t, 1, invocations, "Function should be called only once")
}

// Custom error type to test panic propagation
type customError struct {
	message string
}

func (e customError) Error() string {
	return e.message
}

func testPanicPropagation(t *testing.T) {
	memoizer := NewMemoizer[int]()

	customErr := customError{message: "custom panic error"}

	assert.PanicsWithValue(t, customErr, func() {
		_, _ = memoizer.Memoize("panic_key", func() (int, error) {
			panic(customErr)
		})
	}, "Memoizer should propagate the original panic value")
}

func testNoMemoizationOnError(t *testing.T) {
	memoizer := NewMemoizer[int]()
	key := "error_key"
	callCount := 0

	// First call - should return an error
	_, err := memoizer.Memoize(key, func() (int, error) {
		callCount++
		return 0, errors.New("intentional error")
	})
	require.Error(t, err)
	assert.Equal(t, 1, callCount)

	// Second call - should call the function again, not use a cached value
	_, err = memoizer.Memoize(key, func() (int, error) {
		callCount++
		return 0, errors.New("intentional error")
	})
	require.Error(t, err)
	assert.Equal(t, 2, callCount, "Function should be called again on error")

	// Third call - should succeed and cache the result
	result, err := memoizer.Memoize(key, func() (int, error) {
		callCount++
		return 42, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 3, callCount)

	// Fourth call - should use the cached value
	result, err = memoizer.Memoize(key, func() (int, error) {
		callCount++
		return 0, errors.New("this should not be called")
	})
	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 3, callCount, "Cached value should be used")
}

func testNoMemoizationOnPanic(t *testing.T) {
	memoizer := NewMemoizer[int]()
	key := "panic_key"
	callCount := 0

	// First call - should panic
	assert.Panics(t, func() {
		memoizer.Memoize(key, func() (int, error) {
			callCount++
			panic("intentional panic")
		})
	})
	assert.Equal(t, 1, callCount)

	// Second call - should call the function again, not use a cached value
	assert.Panics(t, func() {
		memoizer.Memoize(key, func() (int, error) {
			callCount++
			panic("intentional panic")
		})
	})
	assert.Equal(t, 2, callCount, "Function should be called again on panic")

	// Third call - should succeed and cache the result
	result, err := memoizer.Memoize(key, func() (int, error) {
		callCount++
		return 42, nil
	})
	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 3, callCount)

	// Fourth call - should use the cached value
	result, err = memoizer.Memoize(key, func() (int, error) {
		callCount++
		return 0, errors.New("this should not be called")
	})
	require.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.Equal(t, 3, callCount, "Cached value should be used")
}
