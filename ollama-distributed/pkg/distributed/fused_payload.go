package distributed

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// FusedPayloadManager implements fused data-flag payloads
// Combines data and control flags in single messages to reduce round-trips
type FusedPayloadManager struct {
	mu sync.RWMutex

	// Configuration
	config *FusedPayloadConfig

	// Performance metrics
	metrics *FusedPayloadMetrics
}

// FusedPayload represents a payload with data and control flags
type FusedPayload struct {
	// Header
	Version     uint8
	PayloadType PayloadType
	Flags       PayloadFlags
	DataSize    uint64
	Timestamp   int64

	// Data
	Data []byte

	// Metadata
	SequenceNumber uint64
	ChecksumCRC32  uint32
}

// PayloadType defines the type of payload
type PayloadType uint8

const (
	PayloadActivation PayloadType = iota
	PayloadGradient
	PayloadWeight
	PayloadControl
	PayloadHeartbeat
)

// PayloadFlags contains control flags
type PayloadFlags struct {
	IsCompressed    bool
	RequiresAck     bool
	IsLastChunk     bool
	IsPriority      bool
	HasChecksum     bool
	IsEncrypted     bool
	Reserved        uint16 // Reserved for future use
}

// FusedPayloadConfig configures fused payload behavior
type FusedPayloadConfig struct {
	EnableCompression bool
	EnableChecksum    bool
	EnableEncryption  bool
	MaxPayloadSize    uint64
}

// FusedPayloadMetrics tracks performance metrics
type FusedPayloadMetrics struct {
	mu                    sync.RWMutex
	TotalPayloads         int64
	TotalDataBytes        uint64
	TotalOverheadBytes    uint64
	AverageOverheadRatio  float64
	RoundTripsReduced     int64
	CompressionRatio      float64
}

// DefaultFusedPayloadConfig returns default configuration
func DefaultFusedPayloadConfig() *FusedPayloadConfig {
	return &FusedPayloadConfig{
		EnableCompression: true,
		EnableChecksum:    true,
		EnableEncryption:  false,
		MaxPayloadSize:    10 * 1024 * 1024, // 10MB
	}
}

// NewFusedPayloadManager creates a new fused payload manager
func NewFusedPayloadManager(config *FusedPayloadConfig) *FusedPayloadManager {
	if config == nil {
		config = DefaultFusedPayloadConfig()
	}

	return &FusedPayloadManager{
		config:  config,
		metrics: &FusedPayloadMetrics{},
	}
}

// CreatePayload creates a fused payload with data and flags
func (fpm *FusedPayloadManager) CreatePayload(
	payloadType PayloadType,
	data []byte,
	flags PayloadFlags,
) (*FusedPayload, error) {
	if uint64(len(data)) > fpm.config.MaxPayloadSize {
		return nil, fmt.Errorf("data size %d exceeds max payload size %d", len(data), fpm.config.MaxPayloadSize)
	}

	payload := &FusedPayload{
		Version:     1,
		PayloadType: payloadType,
		Flags:       flags,
		DataSize:    uint64(len(data)),
		Timestamp:   time.Now().UnixNano(),
		Data:        data,
	}

	// Apply compression if enabled
	if fpm.config.EnableCompression && flags.IsCompressed {
		compressed, err := fpm.compressData(data)
		if err == nil && len(compressed) < len(data) {
			payload.Data = compressed
			payload.DataSize = uint64(len(compressed))
		}
	}

	// Calculate checksum if enabled
	if fpm.config.EnableChecksum && flags.HasChecksum {
		payload.ChecksumCRC32 = fpm.calculateCRC32(payload.Data)
	}

	// Update metrics
	fpm.updateMetrics(payload, len(data))

	return payload, nil
}

// Serialize serializes a fused payload to bytes
func (fpm *FusedPayloadManager) Serialize(payload *FusedPayload) ([]byte, error) {
	// Calculate total size
	headerSize := 32 // Fixed header size
	totalSize := headerSize + len(payload.Data)

	buf := make([]byte, totalSize)
	offset := 0

	// Write header
	buf[offset] = payload.Version
	offset++

	buf[offset] = uint8(payload.PayloadType)
	offset++

	// Write flags (2 bytes)
	flagsBytes := fpm.serializeFlags(payload.Flags)
	copy(buf[offset:offset+2], flagsBytes)
	offset += 2

	// Write data size (8 bytes)
	binary.LittleEndian.PutUint64(buf[offset:offset+8], payload.DataSize)
	offset += 8

	// Write timestamp (8 bytes)
	binary.LittleEndian.PutUint64(buf[offset:offset+8], uint64(payload.Timestamp))
	offset += 8

	// Write sequence number (8 bytes)
	binary.LittleEndian.PutUint64(buf[offset:offset+8], payload.SequenceNumber)
	offset += 8

	// Write checksum (4 bytes)
	binary.LittleEndian.PutUint32(buf[offset:offset+4], payload.ChecksumCRC32)
	offset += 4

	// Write data
	copy(buf[offset:], payload.Data)

	return buf, nil
}

// Deserialize deserializes bytes to a fused payload
func (fpm *FusedPayloadManager) Deserialize(data []byte) (*FusedPayload, error) {
	if len(data) < 32 {
		return nil, fmt.Errorf("data too short for header")
	}

	payload := &FusedPayload{}
	offset := 0

	// Read header
	payload.Version = data[offset]
	offset++

	payload.PayloadType = PayloadType(data[offset])
	offset++

	// Read flags
	payload.Flags = fpm.deserializeFlags(data[offset : offset+2])
	offset += 2

	// Read data size
	payload.DataSize = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Read timestamp
	payload.Timestamp = int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8

	// Read sequence number
	payload.SequenceNumber = binary.LittleEndian.Uint64(data[offset : offset+8])
	offset += 8

	// Read checksum
	payload.ChecksumCRC32 = binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read data
	payload.Data = make([]byte, payload.DataSize)
	copy(payload.Data, data[offset:])

	// Verify checksum if enabled
	if payload.Flags.HasChecksum {
		calculatedChecksum := fpm.calculateCRC32(payload.Data)
		if calculatedChecksum != payload.ChecksumCRC32 {
			return nil, fmt.Errorf("checksum mismatch")
		}
	}

	return payload, nil
}

// serializeFlags serializes flags to bytes
func (fpm *FusedPayloadManager) serializeFlags(flags PayloadFlags) []byte {
	var flagByte uint16 = 0
	if flags.IsCompressed {
		flagByte |= 1 << 0
	}
	if flags.RequiresAck {
		flagByte |= 1 << 1
	}
	if flags.IsLastChunk {
		flagByte |= 1 << 2
	}
	if flags.IsPriority {
		flagByte |= 1 << 3
	}
	if flags.HasChecksum {
		flagByte |= 1 << 4
	}
	if flags.IsEncrypted {
		flagByte |= 1 << 5
	}

	buf := make([]byte, 2)
	binary.LittleEndian.PutUint16(buf, flagByte)
	return buf
}

// deserializeFlags deserializes bytes to flags
func (fpm *FusedPayloadManager) deserializeFlags(data []byte) PayloadFlags {
	flagByte := binary.LittleEndian.Uint16(data)
	return PayloadFlags{
		IsCompressed: (flagByte & (1 << 0)) != 0,
		RequiresAck:  (flagByte & (1 << 1)) != 0,
		IsLastChunk:  (flagByte & (1 << 2)) != 0,
		IsPriority:   (flagByte & (1 << 3)) != 0,
		HasChecksum:  (flagByte & (1 << 4)) != 0,
		IsEncrypted:  (flagByte & (1 << 5)) != 0,
	}
}

// compressData compresses data (placeholder)
func (fpm *FusedPayloadManager) compressData(data []byte) ([]byte, error) {
	// Placeholder - actual implementation would use compression library
	return data, nil
}

// calculateCRC32 calculates CRC32 checksum (placeholder)
func (fpm *FusedPayloadManager) calculateCRC32(data []byte) uint32 {
	// Placeholder - actual implementation would use hash/crc32
	return 0
}

// updateMetrics updates performance metrics
func (fpm *FusedPayloadManager) updateMetrics(payload *FusedPayload, originalSize int) {
	fpm.metrics.mu.Lock()
	defer fpm.metrics.mu.Unlock()

	fpm.metrics.TotalPayloads++
	fpm.metrics.TotalDataBytes += uint64(originalSize)
	fpm.metrics.TotalOverheadBytes += 32 // Header size

	// Calculate overhead ratio
	overheadRatio := float64(32) / float64(originalSize+32)
	alpha := 0.3
	fpm.metrics.AverageOverheadRatio = fpm.metrics.AverageOverheadRatio*(1-alpha) + overheadRatio*alpha

	// Estimate round-trips reduced (fusing data+flags saves 1 round-trip per message)
	fpm.metrics.RoundTripsReduced++

	// Calculate compression ratio if compressed
	if payload.Flags.IsCompressed && len(payload.Data) < originalSize {
		compressionRatio := float64(len(payload.Data)) / float64(originalSize)
		fpm.metrics.CompressionRatio = fpm.metrics.CompressionRatio*(1-alpha) + compressionRatio*alpha
	}
}

// GetMetrics returns current performance metrics
func (fpm *FusedPayloadManager) GetMetrics() *FusedPayloadMetrics {
	fpm.metrics.mu.RLock()
	defer fpm.metrics.mu.RUnlock()

	return &FusedPayloadMetrics{
		TotalPayloads:        fpm.metrics.TotalPayloads,
		TotalDataBytes:       fpm.metrics.TotalDataBytes,
		TotalOverheadBytes:   fpm.metrics.TotalOverheadBytes,
		AverageOverheadRatio: fpm.metrics.AverageOverheadRatio,
		RoundTripsReduced:    fpm.metrics.RoundTripsReduced,
		CompressionRatio:     fpm.metrics.CompressionRatio,
	}
}

