package models

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"sync"
	"time"
)

// IntegrityVerifier handles model integrity verification with multiple hash algorithms
type IntegrityVerifier struct {
	mu sync.RWMutex

	// Verification cache
	verificationCache map[string]*VerificationResult

	// Configuration
	config *VerificationConfig

	// Metrics
	metrics *VerificationMetrics
}

// VerificationResult represents the result of an integrity verification
type VerificationResult struct {
	ModelName    string `json:"model_name"`
	ModelVersion string `json:"model_version"`
	FilePath     string `json:"file_path"`
	FileSize     int64  `json:"file_size"`

	// Hash results
	Checksums         map[HashAlgorithm]string `json:"checksums"`
	ExpectedChecksums map[HashAlgorithm]string `json:"expected_checksums"`

	// Verification status
	Verified         bool          `json:"verified"`
	VerificationTime time.Time     `json:"verification_time"`
	Duration         time.Duration `json:"duration"`

	// Error information
	ErrorMessage     string          `json:"error_message,omitempty"`
	FailedAlgorithms []HashAlgorithm `json:"failed_algorithms,omitempty"`
}

// VerificationConfig configures integrity verification
type VerificationConfig struct {
	// Hash algorithms to use
	HashAlgorithms []HashAlgorithm

	// Verification settings
	EnableCaching bool
	CacheTimeout  time.Duration
	MaxCacheSize  int

	// Performance settings
	BufferSize     int
	EnableParallel bool
	MaxConcurrent  int

	// Verification policies
	RequireAllHashes bool
	FailOnFirstError bool
	EnableDeepVerify bool
}

// VerificationMetrics tracks verification performance
type VerificationMetrics struct {
	TotalVerifications       int64                   `json:"total_verifications"`
	SuccessfulVerifications  int64                   `json:"successful_verifications"`
	FailedVerifications      int64                   `json:"failed_verifications"`
	CacheHits                int64                   `json:"cache_hits"`
	CacheMisses              int64                   `json:"cache_misses"`
	AverageVerificationTime  time.Duration           `json:"average_verification_time"`
	VerificationsByAlgorithm map[HashAlgorithm]int64 `json:"verifications_by_algorithm"`
	LastUpdated              time.Time               `json:"last_updated"`
}

// HashAlgorithm represents supported hash algorithms
type HashAlgorithm string

const (
	HashAlgorithmMD5    HashAlgorithm = "md5"
	HashAlgorithmSHA1   HashAlgorithm = "sha1"
	HashAlgorithmSHA256 HashAlgorithm = "sha256"
	HashAlgorithmSHA512 HashAlgorithm = "sha512"
)

// NewIntegrityVerifier creates a new integrity verifier
func NewIntegrityVerifier(config *VerificationConfig) *IntegrityVerifier {
	if config == nil {
		config = &VerificationConfig{
			HashAlgorithms:   []HashAlgorithm{HashAlgorithmSHA256, HashAlgorithmSHA512},
			EnableCaching:    true,
			CacheTimeout:     time.Hour,
			MaxCacheSize:     1000,
			BufferSize:       64 * 1024, // 64KB buffer
			EnableParallel:   true,
			MaxConcurrent:    4,
			RequireAllHashes: false,
			FailOnFirstError: false,
			EnableDeepVerify: true,
		}
	}

	return &IntegrityVerifier{
		verificationCache: make(map[string]*VerificationResult),
		config:            config,
		metrics: &VerificationMetrics{
			VerificationsByAlgorithm: make(map[HashAlgorithm]int64),
		},
	}
}

// VerifyModel verifies the integrity of a model file
func (iv *IntegrityVerifier) VerifyModel(modelName, modelVersion, filePath string, expectedChecksums map[HashAlgorithm]string) (*VerificationResult, error) {
	startTime := time.Now()

	// Check cache first
	if iv.config.EnableCaching {
		cacheKey := fmt.Sprintf("%s_%s_%s", modelName, modelVersion, filePath)
		if result := iv.getCachedResult(cacheKey); result != nil {
			iv.metrics.CacheHits++
			return result, nil
		}
		iv.metrics.CacheMisses++
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Create verification result
	result := &VerificationResult{
		ModelName:         modelName,
		ModelVersion:      modelVersion,
		FilePath:          filePath,
		FileSize:          fileInfo.Size(),
		Checksums:         make(map[HashAlgorithm]string),
		ExpectedChecksums: expectedChecksums,
		VerificationTime:  startTime,
	}

	// Calculate checksums
	if err := iv.calculateChecksums(filePath, result); err != nil {
		result.ErrorMessage = err.Error()
		result.Duration = time.Since(startTime)
		iv.metrics.FailedVerifications++
		return result, err
	}

	// Verify checksums
	result.Verified = iv.verifyChecksums(result)
	result.Duration = time.Since(startTime)

	// Update metrics
	iv.updateMetrics(result)

	// Cache result
	if iv.config.EnableCaching {
		cacheKey := fmt.Sprintf("%s_%s_%s", modelName, modelVersion, filePath)
		iv.cacheResult(cacheKey, result)
	}

	return result, nil
}

// calculateChecksums calculates checksums for all configured algorithms
func (iv *IntegrityVerifier) calculateChecksums(filePath string, result *VerificationResult) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if iv.config.EnableParallel && len(iv.config.HashAlgorithms) > 1 {
		return iv.calculateChecksumsParallel(file, result)
	}

	return iv.calculateChecksumsSequential(file, result)
}

// calculateChecksumsSequential calculates checksums sequentially
func (iv *IntegrityVerifier) calculateChecksumsSequential(file *os.File, result *VerificationResult) error {
	for _, algorithm := range iv.config.HashAlgorithms {
		// Reset file position
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to seek file: %w", err)
		}

		checksum, err := iv.calculateSingleChecksum(file, algorithm)
		if err != nil {
			if iv.config.FailOnFirstError {
				return err
			}
			result.FailedAlgorithms = append(result.FailedAlgorithms, algorithm)
			continue
		}

		result.Checksums[algorithm] = checksum
		iv.metrics.VerificationsByAlgorithm[algorithm]++
	}

	return nil
}

// calculateChecksumsParallel calculates checksums in parallel using multiple readers
func (iv *IntegrityVerifier) calculateChecksumsParallel(file *os.File, result *VerificationResult) error {
	// Create hash writers for each algorithm
	hashers := make(map[HashAlgorithm]hash.Hash)
	writers := make([]io.Writer, 0, len(iv.config.HashAlgorithms))

	for _, algorithm := range iv.config.HashAlgorithms {
		hasher := iv.createHasher(algorithm)
		if hasher == nil {
			continue
		}
		hashers[algorithm] = hasher
		writers = append(writers, hasher)
	}

	if len(writers) == 0 {
		return fmt.Errorf("no valid hash algorithms configured")
	}

	// Create multi-writer to write to all hashers simultaneously
	multiWriter := io.MultiWriter(writers...)

	// Copy file data to all hashers
	buffer := make([]byte, iv.config.BufferSize)
	if _, err := io.CopyBuffer(multiWriter, file, buffer); err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Extract checksums
	for algorithm, hasher := range hashers {
		checksum := hex.EncodeToString(hasher.Sum(nil))
		result.Checksums[algorithm] = checksum
		iv.metrics.VerificationsByAlgorithm[algorithm]++
	}

	return nil
}

// calculateSingleChecksum calculates checksum for a single algorithm
func (iv *IntegrityVerifier) calculateSingleChecksum(reader io.Reader, algorithm HashAlgorithm) (string, error) {
	hasher := iv.createHasher(algorithm)
	if hasher == nil {
		return "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	buffer := make([]byte, iv.config.BufferSize)
	if _, err := io.CopyBuffer(hasher, reader, buffer); err != nil {
		return "", fmt.Errorf("failed to calculate %s checksum: %w", algorithm, err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// createHasher creates a hasher for the specified algorithm
func (iv *IntegrityVerifier) createHasher(algorithm HashAlgorithm) hash.Hash {
	switch algorithm {
	case HashAlgorithmMD5:
		return md5.New()
	case HashAlgorithmSHA1:
		return sha1.New()
	case HashAlgorithmSHA256:
		return sha256.New()
	case HashAlgorithmSHA512:
		return sha512.New()
	default:
		return nil
	}
}

// verifyChecksums verifies calculated checksums against expected values
func (iv *IntegrityVerifier) verifyChecksums(result *VerificationResult) bool {
	if len(result.ExpectedChecksums) == 0 {
		// No expected checksums provided, consider verified if we calculated any
		return len(result.Checksums) > 0
	}

	verified := true
	matchedAlgorithms := 0

	for algorithm, expectedChecksum := range result.ExpectedChecksums {
		actualChecksum, exists := result.Checksums[algorithm]
		if !exists {
			if iv.config.RequireAllHashes {
				verified = false
				result.FailedAlgorithms = append(result.FailedAlgorithms, algorithm)
			}
			continue
		}

		if actualChecksum != expectedChecksum {
			verified = false
			result.FailedAlgorithms = append(result.FailedAlgorithms, algorithm)
		} else {
			matchedAlgorithms++
		}
	}

	// If we don't require all hashes, at least one must match
	if !iv.config.RequireAllHashes && matchedAlgorithms == 0 && len(result.ExpectedChecksums) > 0 {
		verified = false
	}

	return verified
}

// VerifyChunk verifies the integrity of a model chunk
func (iv *IntegrityVerifier) VerifyChunk(chunkData []byte, expectedChecksum string, algorithm HashAlgorithm) (bool, string, error) {
	hasher := iv.createHasher(algorithm)
	if hasher == nil {
		return false, "", fmt.Errorf("unsupported hash algorithm: %s", algorithm)
	}

	hasher.Write(chunkData)
	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	verified := actualChecksum == expectedChecksum
	return verified, actualChecksum, nil
}

// QuickVerify performs a quick verification using only SHA256
func (iv *IntegrityVerifier) QuickVerify(filePath string, expectedSHA256 string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	checksum, err := iv.calculateSingleChecksum(file, HashAlgorithmSHA256)
	if err != nil {
		return false, err
	}

	return checksum == expectedSHA256, nil
}

// DeepVerify performs deep verification with additional checks
func (iv *IntegrityVerifier) DeepVerify(modelName, modelVersion, filePath string, expectedChecksums map[HashAlgorithm]string) (*VerificationResult, error) {
	if !iv.config.EnableDeepVerify {
		return iv.VerifyModel(modelName, modelVersion, filePath, expectedChecksums)
	}

	// Perform standard verification first
	result, err := iv.VerifyModel(modelName, modelVersion, filePath, expectedChecksums)
	if err != nil {
		return result, err
	}

	// Additional deep verification checks
	if result.Verified {
		// Check file permissions
		fileInfo, err := os.Stat(filePath)
		if err == nil {
			// Verify file is readable
			if fileInfo.Mode().Perm()&0400 == 0 {
				result.Verified = false
				result.ErrorMessage = "file is not readable"
			}
		}

		// Verify file is not empty
		if result.FileSize == 0 {
			result.Verified = false
			result.ErrorMessage = "file is empty"
		}

		// Additional format-specific checks could be added here
		// For example, checking if the file is a valid model format
	}

	return result, nil
}

// getCachedResult retrieves a cached verification result
func (iv *IntegrityVerifier) getCachedResult(cacheKey string) *VerificationResult {
	iv.mu.RLock()
	defer iv.mu.RUnlock()

	result, exists := iv.verificationCache[cacheKey]
	if !exists {
		return nil
	}

	// Check if cache entry is still valid
	if time.Since(result.VerificationTime) > iv.config.CacheTimeout {
		delete(iv.verificationCache, cacheKey)
		return nil
	}

	return result
}

// cacheResult caches a verification result
func (iv *IntegrityVerifier) cacheResult(cacheKey string, result *VerificationResult) {
	iv.mu.Lock()
	defer iv.mu.Unlock()

	// Check cache size limit
	if len(iv.verificationCache) >= iv.config.MaxCacheSize {
		// Remove oldest entry
		var oldestKey string
		var oldestTime time.Time

		for key, cachedResult := range iv.verificationCache {
			if oldestKey == "" || cachedResult.VerificationTime.Before(oldestTime) {
				oldestKey = key
				oldestTime = cachedResult.VerificationTime
			}
		}

		if oldestKey != "" {
			delete(iv.verificationCache, oldestKey)
		}
	}

	// Cache the result
	iv.verificationCache[cacheKey] = result
}

// updateMetrics updates verification metrics
func (iv *IntegrityVerifier) updateMetrics(result *VerificationResult) {
	iv.mu.Lock()
	defer iv.mu.Unlock()

	iv.metrics.TotalVerifications++

	if result.Verified {
		iv.metrics.SuccessfulVerifications++
	} else {
		iv.metrics.FailedVerifications++
	}

	// Update average verification time
	if iv.metrics.TotalVerifications == 1 {
		iv.metrics.AverageVerificationTime = result.Duration
	} else {
		totalTime := time.Duration(iv.metrics.TotalVerifications-1)*iv.metrics.AverageVerificationTime + result.Duration
		iv.metrics.AverageVerificationTime = totalTime / time.Duration(iv.metrics.TotalVerifications)
	}

	iv.metrics.LastUpdated = time.Now()
}

// GetMetrics returns verification metrics
func (iv *IntegrityVerifier) GetMetrics() *VerificationMetrics {
	iv.mu.RLock()
	defer iv.mu.RUnlock()

	metrics := *iv.metrics
	return &metrics
}

// ClearCache clears the verification cache
func (iv *IntegrityVerifier) ClearCache() {
	iv.mu.Lock()
	defer iv.mu.Unlock()

	iv.verificationCache = make(map[string]*VerificationResult)
}

// GetCacheStats returns cache statistics
func (iv *IntegrityVerifier) GetCacheStats() map[string]interface{} {
	iv.mu.RLock()
	defer iv.mu.RUnlock()

	return map[string]interface{}{
		"cache_size":     len(iv.verificationCache),
		"max_cache_size": iv.config.MaxCacheSize,
		"cache_hits":     iv.metrics.CacheHits,
		"cache_misses":   iv.metrics.CacheMisses,
		"hit_ratio":      float64(iv.metrics.CacheHits) / float64(iv.metrics.CacheHits+iv.metrics.CacheMisses),
	}
}
