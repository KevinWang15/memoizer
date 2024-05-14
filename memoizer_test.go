package memoizer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoizer_Memoize(t *testing.T) {
	memoizer := NewMemoizer[int]()

	// Test case 1: Successful memoization
	result1, err1 := memoizer.Memoize("key1", func() (int, error) {
		return 42, nil
	})
	assert.NoError(t, err1)
	assert.Equal(t, 42, result1)

	// Test case 2: Cached result retrieval
	result2, err2 := memoizer.Memoize("key1", func() (int, error) {
		return 0, errors.New("error occurred")
	})
	assert.NoError(t, err2)
	assert.Equal(t, 42, result2)

	// Test case 3: Error handling
	_, err3 := memoizer.Memoize("key2", func() (int, error) {
		return 0, errors.New("error occurred")
	})
	assert.Error(t, err3)

	// Test case 4: Concurrent memoization
	resultChan := make(chan int)
	errChan := make(chan error)
	actualInvocations := 0

	for i := 0; i < 5; i++ {
		go func() {
			result, err := memoizer.Memoize("key3", func() (int, error) {
				actualInvocations++
				return 100, nil
			})
			resultChan <- result
			errChan <- err
		}()
	}

	for i := 0; i < 5; i++ {
		result := <-resultChan
		err := <-errChan
		assert.NoError(t, err)
		assert.Equal(t, 100, result)
		assert.Equal(t, 1, actualInvocations)
	}
}
