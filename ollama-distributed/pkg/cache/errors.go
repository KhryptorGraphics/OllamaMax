package cache

import "errors"

// Common cache errors
var (
	ErrCacheNotFound = errors.New("cache not found")
	ErrKeyTooLarge   = errors.New("key size exceeds maximum")
	ErrValueTooLarge = errors.New("value size exceeds maximum")
	ErrCacheClosed   = errors.New("cache is closed")
)