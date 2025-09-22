package messaging

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sync"
	"time"
	"unsafe"

	"github.com/klauspost/compress/zstd"
)

// JSONSerializer implements JSON-based message serialization with fragmentation support
type JSONSerializer struct {
	enableCompression bool
	compressionLevel  int
	maxMessageSize    int // Maximum message size before fragmentation
	fragmentManager   *FragmentManager
}

// BinarySerializer implements binary message serialization
type BinarySerializer struct {
	enableCompression bool
	compressionLevel  int
}

// FragmentManager handles message fragment reassembly
type FragmentManager struct {
	fragments    map[string]*PartialMessage // fragment_id -> partial message
	timeouts     map[string]time.Time       // fragment_id -> expiry time
	mutex        sync.RWMutex
	cleanupTicker *time.Ticker
	done         chan struct{}
}

// PartialMessage represents a partially assembled fragmented message
type PartialMessage struct {
	Fragments      map[uint16]*MessageFrame // fragment_index -> frame
	TotalFragments uint16
	LastActivity   time.Time
	Complete       bool
}

// MessageFrame represents a framed message for transmission
type MessageFrame struct {
	Version     uint8  `json:"version"`
	MessageType uint8  `json:"message_type"`
	Flags       uint8  `json:"flags"`
	Length      uint32 `json:"length"`
	Checksum    uint32 `json:"checksum"`
	Data        []byte `json:"data"`

	// Fragmentation metadata (only set when FlagFragmented is true)
	FragmentID    string `json:"fragment_id,omitempty"`    // Unique ID for fragment group
	FragmentIndex uint16 `json:"fragment_index,omitempty"` // 0-based fragment index
	TotalFragments uint16 `json:"total_fragments,omitempty"` // Total number of fragments
}

// Serialization flags
const (
	FlagCompressed = 1 << iota
	FlagEncrypted
	FlagFragmented
	FlagTensorData
	FlagQuantized
	FlagSparse
)

// Message frame version
const (
	FrameVersion1 = 1
)

// Fragmentation constants
const (
	DefaultFragmentSize      = 8 * 1024 * 1024  // 8MB fragments (smaller than 10MB router limit)
	FragmentTimeout          = 30 * time.Second // Fragment reassembly timeout
	FragmentCleanupInterval  = 5 * time.Minute  // How often to cleanup expired fragments
	MaxFragmentsPerMessage   = 1000             // Maximum fragments per message
)

// NewJSONSerializer creates a new JSON serializer with fragmentation support
func NewJSONSerializer(enableCompression bool) *JSONSerializer {
	js := &JSONSerializer{
		enableCompression: enableCompression,
		compressionLevel:  gzip.DefaultCompression,
		maxMessageSize:    DefaultFragmentSize,
		fragmentManager:   NewFragmentManager(),
	}
	return js
}

// NewFragmentManager creates a new fragment manager
func NewFragmentManager() *FragmentManager {
	fm := &FragmentManager{
		fragments:     make(map[string]*PartialMessage),
		timeouts:      make(map[string]time.Time),
		cleanupTicker: time.NewTicker(FragmentCleanupInterval),
		done:          make(chan struct{}),
	}

	// Start cleanup routine
	go fm.cleanupRoutine()

	return fm
}

// Serialize serializes a message to bytes, with fragmentation support for large messages
func (js *JSONSerializer) Serialize(msg *Message) ([]byte, error) {
	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Compress if enabled and beneficial
	flags := uint8(0)
	if js.enableCompression && len(data) > 1024 {
		compressed, err := js.compress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to compress message: %w", err)
		}

		// Use compressed data if it's smaller
		if len(compressed) < len(data) {
			data = compressed
			flags |= FlagCompressed
		}
	}

	// Check if fragmentation is needed
	if len(data) > js.maxMessageSize {
		return js.serializeFragmented(msg, data, flags)
	}

	// Create single frame for small messages
	frame := &MessageFrame{
		Version:     FrameVersion1,
		MessageType: uint8(getMessageTypeCode(msg.Type)),
		Flags:       flags,
		Length:      uint32(len(data)),
		Checksum:    calculateChecksum(data),
		Data:        data,
	}

	return js.serializeFrame(frame)
}

// serializeFragmented creates multiple fragments for large messages
func (js *JSONSerializer) serializeFragmented(msg *Message, data []byte, baseFlags uint8) ([]byte, error) {
	// Generate unique fragment ID
	fragmentID, err := generateFragmentID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate fragment ID: %w", err)
	}

	// Calculate number of fragments needed
	totalFragments := uint16((len(data) + js.maxMessageSize - 1) / js.maxMessageSize)
	if totalFragments > MaxFragmentsPerMessage {
		return nil, fmt.Errorf("message too large: requires %d fragments (max %d)", totalFragments, MaxFragmentsPerMessage)
	}

	// Create first fragment with complete metadata
	firstChunk := data[:min(len(data), js.maxMessageSize)]
	firstFrame := &MessageFrame{
		Version:        FrameVersion1,
		MessageType:    uint8(getMessageTypeCode(msg.Type)),
		Flags:          baseFlags | FlagFragmented,
		Length:         uint32(len(firstChunk)),
		Checksum:       calculateChecksum(firstChunk),
		Data:           firstChunk,
		FragmentID:     fragmentID,
		FragmentIndex:  0,
		TotalFragments: totalFragments,
	}

	return js.serializeFrame(firstFrame)
}

// generateFragmentID generates a unique fragment ID
func generateFragmentID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Deserialize deserializes bytes to a message, handling fragment reassembly
func (js *JSONSerializer) Deserialize(data []byte) (*Message, error) {
	// Deserialize frame
	frame, err := js.deserializeFrame(data)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize frame: %w", err)
	}

	// Verify checksum
	if frame.Checksum != calculateChecksum(frame.Data) {
		return nil, fmt.Errorf("checksum mismatch")
	}

	// Handle fragmented messages
	if frame.Flags&FlagFragmented != 0 {
		return js.handleFragmentedMessage(frame)
	}

	// Handle non-fragmented messages
	messageData := frame.Data
	if frame.Flags&FlagCompressed != 0 {
		decompressed, err := js.decompress(frame.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress message: %w", err)
		}
		messageData = decompressed
	}

	// Unmarshal message
	var msg Message
	if err := json.Unmarshal(messageData, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// handleFragmentedMessage processes a fragment and returns complete message if available
func (js *JSONSerializer) handleFragmentedMessage(frame *MessageFrame) (*Message, error) {
	// Add fragment to reassembly manager
	completeData, err := js.fragmentManager.AddFragment(frame)
	if err != nil {
		return nil, fmt.Errorf("failed to add fragment: %w", err)
	}

	// Return nil if message is not yet complete
	if completeData == nil {
		return nil, nil
	}

	// Process complete reassembled data
	messageData := completeData
	if frame.Flags&FlagCompressed != 0 {
		decompressed, err := js.decompress(completeData)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress reassembled message: %w", err)
		}
		messageData = decompressed
	}

	// Unmarshal complete message
	var msg Message
	if err := json.Unmarshal(messageData, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal reassembled message: %w", err)
	}

	return &msg, nil
}

// compress compresses data using gzip
func (js *JSONSerializer) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, js.compressionLevel)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompress decompresses gzip data
func (js *JSONSerializer) decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// serializeFrame serializes a message frame
func (js *JSONSerializer) serializeFrame(frame *MessageFrame) ([]byte, error) {
	return json.Marshal(frame)
}

// deserializeFrame deserializes a message frame
func (js *JSONSerializer) deserializeFrame(data []byte) (*MessageFrame, error) {
	var frame MessageFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		return nil, err
	}
	return &frame, nil
}

// NewBinarySerializer creates a new binary serializer
func NewBinarySerializer(enableCompression bool) *BinarySerializer {
	return &BinarySerializer{
		enableCompression: enableCompression,
		compressionLevel:  gzip.DefaultCompression,
	}
}

// Serialize serializes a message to bytes using binary format
func (bs *BinarySerializer) Serialize(msg *Message) ([]byte, error) {
	// For now, use JSON serialization as binary format is more complex
	// In a production system, this would use a more efficient binary format
	// like Protocol Buffers, MessagePack, or custom binary encoding

	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Compress if enabled and beneficial
	if bs.enableCompression && len(data) > 512 {
		compressed, err := bs.compress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to compress message: %w", err)
		}

		// Use compressed data if it's smaller
		if len(compressed) < len(data) {
			return bs.createBinaryFrame(compressed, FlagCompressed)
		}
	}

	return bs.createBinaryFrame(data, 0)
}

// Deserialize deserializes bytes to a message using binary format
func (bs *BinarySerializer) Deserialize(data []byte) (*Message, error) {
	// Parse binary frame
	if len(data) < 11 { // Minimum frame size (fixed off-by-one)
		return nil, fmt.Errorf("data too short for binary frame")
	}

	// Extract frame header (simplified binary format)
	version := data[0]
	if version != FrameVersion1 {
		return nil, fmt.Errorf("unsupported frame version: %d", version)
	}

	flags := data[2]
	length := uint32(data[3])<<24 | uint32(data[4])<<16 | uint32(data[5])<<8 | uint32(data[6])
	checksum := uint32(data[7])<<24 | uint32(data[8])<<16 | uint32(data[9])<<8 | uint32(data[10])

	// Validate bounds after reading length (fixed bounds check)
	if len(data) < int(11+length) {
		return nil, fmt.Errorf("data too short for frame length")
	}

	frameData := data[11 : 11+length]

	// Verify checksum
	if checksum != calculateChecksum(frameData) {
		return nil, fmt.Errorf("checksum mismatch")
	}

	// Decompress if needed
	messageData := frameData
	if flags&FlagCompressed != 0 {
		decompressed, err := bs.decompress(frameData)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress message: %w", err)
		}
		messageData = decompressed
	}

	// Unmarshal message
	var msg Message
	if err := json.Unmarshal(messageData, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// createBinaryFrame creates a binary frame
func (bs *BinarySerializer) createBinaryFrame(data []byte, flags uint8) ([]byte, error) {
	checksum := calculateChecksum(data)
	length := uint32(len(data))

	// Create binary frame (simplified format)
	frame := make([]byte, 11+len(data))
	frame[0] = FrameVersion1 // Version
	frame[1] = 0             // Message type (simplified)
	frame[2] = flags         // Flags
	frame[3] = byte(length >> 24)
	frame[4] = byte(length >> 16)
	frame[5] = byte(length >> 8)
	frame[6] = byte(length)
	frame[7] = byte(checksum >> 24)
	frame[8] = byte(checksum >> 16)
	frame[9] = byte(checksum >> 8)
	frame[10] = byte(checksum)

	copy(frame[11:], data)

	return frame, nil
}

// compress compresses data using gzip
func (bs *BinarySerializer) compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, bs.compressionLevel)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompress decompresses gzip data
func (bs *BinarySerializer) decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// Helper functions

// getMessageTypeCode returns a numeric code for a message type
func getMessageTypeCode(msgType MessageType) int {
	switch msgType {
	case MessageTypeConsensus:
		return 1
	case MessageTypeScheduler:
		return 2
	case MessageTypeModel:
		return 3
	case MessageTypeDiscovery:
		return 4
	case MessageTypeHealth:
		return 5
	case MessageTypeData:
		return 6
	case MessageTypeControl:
		return 7
	case MessageTypeAck:
		return 8
	case MessageTypeTensorStart:
		return 9
	case MessageTypeTensorChunk:
		return 10
	case MessageTypeTensorComplete:
		return 11
	case MessageTypeTensorStream:
		return 12
	default:
		return 0
	}
}

// calculateChecksum calculates a simple checksum for data
func calculateChecksum(data []byte) uint32 {
	var checksum uint32
	for _, b := range data {
		checksum = checksum*31 + uint32(b)
	}
	return checksum
}

// TensorDType represents tensor data types
type TensorDType string

const (
	DTypeFloat32 TensorDType = "float32"
	DTypeFloat16 TensorDType = "float16"
	DTypeBFloat16 TensorDType = "bfloat16"
	DTypeInt8    TensorDType = "int8"
	DTypeInt16   TensorDType = "int16"
	DTypeInt32   TensorDType = "int32"
	DTypeInt64   TensorDType = "int64"
	DTypeUInt8   TensorDType = "uint8"
	DTypeUInt16  TensorDType = "uint16"
	DTypeUInt32  TensorDType = "uint32"
	DTypeUInt64  TensorDType = "uint64"
	DTypeBool    TensorDType = "bool"
)

// TensorMetadata contains tensor-specific metadata
type TensorMetadata struct {
	DType       TensorDType `json:"dtype"`
	Shape       []int64     `json:"shape"`
	Strides     []int64     `json:"strides"`
	Layout      string      `json:"layout"` // "row_major", "col_major"
	Sparse      bool        `json:"sparse"`
	Quantized   bool        `json:"quantized"`
	Scale       float32     `json:"scale,omitempty"`
	ZeroPoint   int32       `json:"zero_point,omitempty"`
	SparsityPattern string  `json:"sparsity_pattern,omitempty"`
}

// TensorSerializer handles efficient tensor data serialization
type TensorSerializer struct {
	enableCompression bool
	compressionLevel  int
	zstdEncoder      *zstd.Encoder
	zstdDecoder      *zstd.Decoder
	quantizationBits int
	sparsityThreshold float32
}

// SerializerFactory creates serializers based on configuration
type SerializerFactory struct {
	defaultFormat     SerializationFormat
	enableCompression bool
	tensorOptimized   bool
}

type SerializationFormat string

const (
	FormatJSON   SerializationFormat = "json"
	FormatBinary SerializationFormat = "binary"
	FormatTensor SerializationFormat = "tensor"
)

// NewTensorSerializer creates a new tensor-optimized serializer
func NewTensorSerializer(enableCompression bool) (*TensorSerializer, error) {
	ts := &TensorSerializer{
		enableCompression: enableCompression,
		compressionLevel:  3, // Fast compression for real-time streaming
		quantizationBits:  8, // Default 8-bit quantization
		sparsityThreshold: 0.9, // 90% sparsity threshold
	}

	if enableCompression {
		// Initialize zstd encoder/decoder for better compression than gzip
		encoder, err := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd encoder: %w", err)
		}
		ts.zstdEncoder = encoder

		decoder, err := zstd.NewReader(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create zstd decoder: %w", err)
		}
		ts.zstdDecoder = decoder
	}

	return ts, nil
}

// SerializeTensor serializes tensor data with optimizations
func (ts *TensorSerializer) SerializeTensor(data []byte, metadata *TensorMetadata) ([]byte, error) {
	var flags uint8 = FlagTensorData
	var processedData []byte = data
	var err error

	// Apply quantization if beneficial
	if ts.shouldQuantize(metadata) {
		processedData, err = ts.quantizeTensor(data, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to quantize tensor: %w", err)
		}
		flags |= FlagQuantized
		metadata.Quantized = true
	}

	// Apply sparse encoding if beneficial
	if ts.shouldApplySparseEncoding(data, metadata) {
		processedData, err = ts.encodeSparse(processedData, metadata)
		if err != nil {
			return nil, fmt.Errorf("failed to encode sparse tensor: %w", err)
		}
		flags |= FlagSparse
		metadata.Sparse = true
	}

	// Compress if enabled and beneficial
	if ts.enableCompression && len(processedData) > 1024 {
		compressed, err := ts.compressTensor(processedData)
		if err != nil {
			return nil, fmt.Errorf("failed to compress tensor: %w", err)
		}

		// Use compression if it reduces size by at least 10%
		if len(compressed) < int(float32(len(processedData))*0.9) {
			processedData = compressed
			flags |= FlagCompressed
		}
	}

	// Create tensor frame
	return ts.createTensorFrame(processedData, metadata, flags)
}

// DeserializeTensor deserializes tensor data with optimizations
func (ts *TensorSerializer) DeserializeTensor(data []byte) ([]byte, *TensorMetadata, error) {
	// Parse tensor frame
	tensorData, metadata, flags, err := ts.parseTensorFrame(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse tensor frame: %w", err)
	}

	processedData := tensorData

	// Decompress if needed
	if flags&FlagCompressed != 0 {
		decompressed, err := ts.decompressTensor(processedData)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decompress tensor: %w", err)
		}
		processedData = decompressed
	}

	// Decode sparse if needed
	if flags&FlagSparse != 0 {
		dense, err := ts.decodeSparse(processedData, metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to decode sparse tensor: %w", err)
		}
		processedData = dense
	}

	// Dequantize if needed
	if flags&FlagQuantized != 0 {
		dequantized, err := ts.dequantizeTensor(processedData, metadata)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to dequantize tensor: %w", err)
		}
		processedData = dequantized
	}

	return processedData, metadata, nil
}

// shouldQuantize determines if quantization should be applied
func (ts *TensorSerializer) shouldQuantize(metadata *TensorMetadata) bool {
	// Apply quantization for float32 tensors in inference (not training)
	return metadata.DType == DTypeFloat32 && !metadata.Quantized
}

// shouldApplySparseEncoding determines if sparse encoding should be applied
func (ts *TensorSerializer) shouldApplySparseEncoding(data []byte, metadata *TensorMetadata) bool {
	if metadata.Sparse {
		return false // Already sparse
	}

	// Calculate sparsity for float32 tensors
	if metadata.DType == DTypeFloat32 {
		sparsity := ts.calculateSparsity(data, metadata)
		return sparsity > ts.sparsityThreshold
	}

	return false
}

// calculateSparsity calculates the sparsity ratio of a tensor
func (ts *TensorSerializer) calculateSparsity(data []byte, metadata *TensorMetadata) float32 {
	if metadata.DType != DTypeFloat32 {
		return 0.0
	}

	floatData := (*[]float32)(unsafe.Pointer(&data))
	totalElements := len(*floatData)
	zeroCount := 0

	for _, v := range *floatData {
		if math.Abs(float64(v)) < 1e-6 {
			zeroCount++
		}
	}

	return float32(zeroCount) / float32(totalElements)
}

// quantizeTensor applies 8-bit quantization to float32 tensors
func (ts *TensorSerializer) quantizeTensor(data []byte, metadata *TensorMetadata) ([]byte, error) {
	if metadata.DType != DTypeFloat32 {
		return data, nil // No quantization for non-float32
	}

	floatData := (*[]float32)(unsafe.Pointer(&data))

	// Find min/max for quantization scale
	var min, max float32 = math.MaxFloat32, -math.MaxFloat32
	for _, v := range *floatData {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// Calculate scale and zero point
	scale := (max - min) / 255.0
	zeroPoint := int32(-min / scale)

	// Quantize to uint8
	quantizedData := make([]uint8, len(*floatData))
	for i, v := range *floatData {
		quantized := int32(v/scale) + zeroPoint
		if quantized < 0 {
			quantized = 0
		} else if quantized > 255 {
			quantized = 255
		}
		quantizedData[i] = uint8(quantized)
	}

	// Store quantization parameters
	metadata.Scale = scale
	metadata.ZeroPoint = zeroPoint

	return (*(*[]byte)(unsafe.Pointer(&quantizedData)))[:len(quantizedData)], nil
}

// dequantizeTensor reverses quantization
func (ts *TensorSerializer) dequantizeTensor(data []byte, metadata *TensorMetadata) ([]byte, error) {
	if !metadata.Quantized {
		return data, nil
	}

	quantizedData := (*[]uint8)(unsafe.Pointer(&data))
	floatData := make([]float32, len(*quantizedData))

	for i, q := range *quantizedData {
		floatData[i] = (float32(q) - float32(metadata.ZeroPoint)) * metadata.Scale
	}

	return (*(*[]byte)(unsafe.Pointer(&floatData)))[:len(floatData)*4], nil
}

// encodeSparse encodes tensor as sparse format (COO - Coordinate format)
func (ts *TensorSerializer) encodeSparse(data []byte, metadata *TensorMetadata) ([]byte, error) {
	if metadata.DType != DTypeFloat32 {
		return data, nil // Only support float32 for now
	}

	floatData := (*[]float32)(unsafe.Pointer(&data))
	var indices []int32
	var values []float32

	// Extract non-zero elements
	for i, v := range *floatData {
		if math.Abs(float64(v)) > 1e-6 {
			indices = append(indices, int32(i))
			values = append(values, v)
		}
	}

	// Create sparse encoding
	buf := bytes.NewBuffer(nil)

	// Write number of non-zero elements
	binary.Write(buf, binary.LittleEndian, int32(len(indices)))

	// Write indices
	for _, idx := range indices {
		binary.Write(buf, binary.LittleEndian, idx)
	}

	// Write values
	for _, val := range values {
		binary.Write(buf, binary.LittleEndian, val)
	}

	metadata.SparsityPattern = "COO"
	return buf.Bytes(), nil
}

// decodeSparse decodes sparse tensor data
func (ts *TensorSerializer) decodeSparse(data []byte, metadata *TensorMetadata) ([]byte, error) {
	if !metadata.Sparse || metadata.SparsityPattern != "COO" {
		return data, nil
	}

	buf := bytes.NewReader(data)

	// Read number of non-zero elements
	var nnz int32
	binary.Read(buf, binary.LittleEndian, &nnz)

	// Read indices
	indices := make([]int32, nnz)
	for i := range indices {
		binary.Read(buf, binary.LittleEndian, &indices[i])
	}

	// Read values
	values := make([]float32, nnz)
	for i := range values {
		binary.Read(buf, binary.LittleEndian, &values[i])
	}

	// Calculate total elements from shape
	totalElements := int32(1)
	for _, dim := range metadata.Shape {
		totalElements *= int32(dim)
	}

	// Reconstruct dense tensor
	denseData := make([]float32, totalElements)
	for i, idx := range indices {
		if idx < totalElements {
			denseData[idx] = values[i]
		}
	}

	return (*(*[]byte)(unsafe.Pointer(&denseData)))[:len(denseData)*4], nil
}

// compressTensor compresses tensor data using zstd
func (ts *TensorSerializer) compressTensor(data []byte) ([]byte, error) {
	if ts.zstdEncoder == nil {
		return ts.compressGzip(data) // Fallback to gzip
	}

	return ts.zstdEncoder.EncodeAll(data, nil), nil
}

// decompressTensor decompresses tensor data
func (ts *TensorSerializer) decompressTensor(data []byte) ([]byte, error) {
	if ts.zstdDecoder == nil {
		return ts.decompressGzip(data) // Fallback to gzip
	}

	return ts.zstdDecoder.DecodeAll(data, nil)
}

// compressGzip fallback compression using gzip
func (ts *TensorSerializer) compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(data); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressGzip fallback decompression using gzip
func (ts *TensorSerializer) decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

// createTensorFrame creates a tensor-specific frame format
func (ts *TensorSerializer) createTensorFrame(data []byte, metadata *TensorMetadata, flags uint8) ([]byte, error) {
	// Serialize metadata
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Calculate checksums
	dataChecksum := calculateChecksum(data)
	metadataChecksum := calculateChecksum(metadataBytes)

	// Create tensor frame header
	buf := bytes.NewBuffer(nil)

	// Frame header (24 bytes)
	binary.Write(buf, binary.LittleEndian, uint8(FrameVersion1)) // Version (1 byte)
	binary.Write(buf, binary.LittleEndian, uint8(255))           // Tensor frame marker (1 byte)
	binary.Write(buf, binary.LittleEndian, flags)                // Flags (1 byte)
	binary.Write(buf, binary.LittleEndian, uint8(0))             // Reserved (1 byte)
	binary.Write(buf, binary.LittleEndian, uint32(len(data)))    // Data length (4 bytes)
	binary.Write(buf, binary.LittleEndian, uint32(len(metadataBytes))) // Metadata length (4 bytes)
	binary.Write(buf, binary.LittleEndian, dataChecksum)         // Data checksum (4 bytes)
	binary.Write(buf, binary.LittleEndian, metadataChecksum)     // Metadata checksum (4 bytes)
	binary.Write(buf, binary.LittleEndian, uint32(0))            // Padding (4 bytes)

	// Write metadata
	buf.Write(metadataBytes)

	// Write data
	buf.Write(data)

	return buf.Bytes(), nil
}

// parseTensorFrame parses a tensor frame
func (ts *TensorSerializer) parseTensorFrame(data []byte) ([]byte, *TensorMetadata, uint8, error) {
	if len(data) < 24 {
		return nil, nil, 0, fmt.Errorf("data too short for tensor frame")
	}

	buf := bytes.NewReader(data)

	// Read frame header
	var version, marker, flags, reserved uint8
	var dataLen, metadataLen, dataChecksum, metadataChecksum, padding uint32

	binary.Read(buf, binary.LittleEndian, &version)
	binary.Read(buf, binary.LittleEndian, &marker)
	binary.Read(buf, binary.LittleEndian, &flags)
	binary.Read(buf, binary.LittleEndian, &reserved)
	binary.Read(buf, binary.LittleEndian, &dataLen)
	binary.Read(buf, binary.LittleEndian, &metadataLen)
	binary.Read(buf, binary.LittleEndian, &dataChecksum)
	binary.Read(buf, binary.LittleEndian, &metadataChecksum)
	binary.Read(buf, binary.LittleEndian, &padding)

	if version != FrameVersion1 || marker != 255 {
		return nil, nil, 0, fmt.Errorf("invalid tensor frame header")
	}

	if len(data) < int(24+metadataLen+dataLen) {
		return nil, nil, 0, fmt.Errorf("data too short for frame content")
	}

	// Read metadata
	metadataBytes := data[24 : 24+metadataLen]
	if calculateChecksum(metadataBytes) != metadataChecksum {
		return nil, nil, 0, fmt.Errorf("metadata checksum mismatch")
	}

	var metadata TensorMetadata
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, nil, 0, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Read tensor data
	tensorData := data[24+metadataLen : 24+metadataLen+dataLen]
	if calculateChecksum(tensorData) != dataChecksum {
		return nil, nil, 0, fmt.Errorf("tensor data checksum mismatch")
	}

	return tensorData, &metadata, flags, nil
}

// Serialize implements MessageSerializer interface for tensor-optimized serialization
func (ts *TensorSerializer) Serialize(msg *Message) ([]byte, error) {
	// Check if this is a tensor message
	if ts.isTensorMessage(msg) {
		return ts.serializeTensorMessage(msg)
	}

	// Fall back to JSON serialization for non-tensor messages
	jsonSerializer := NewJSONSerializer(ts.enableCompression)
	return jsonSerializer.Serialize(msg)
}

// Deserialize implements MessageSerializer interface for tensor-optimized deserialization
func (ts *TensorSerializer) Deserialize(data []byte) (*Message, error) {
	// Check if this is a tensor frame
	if ts.isTensorFrame(data) {
		return ts.deserializeTensorMessage(data)
	}

	// Fall back to JSON deserialization
	jsonSerializer := NewJSONSerializer(ts.enableCompression)
	return jsonSerializer.Deserialize(data)
}

// isTensorMessage checks if a message contains tensor data
func (ts *TensorSerializer) isTensorMessage(msg *Message) bool {
	// Check message type for tensor-related messages
	switch msg.Type {
	case MessageTypeTensorStart, MessageTypeTensorChunk, MessageTypeTensorComplete, MessageTypeTensorStream:
		return true
	default:
		// Check headers for tensor indicators
		if msg.Headers != nil {
			if contentType, ok := msg.Headers["Content-Type"]; ok {
				if contentType == "application/tensor" || contentType == "application/activation" {
					return true
				}
			}
			if _, hasTensorFlag := msg.Headers["X-Tensor-Data"]; hasTensorFlag {
				return true
			}
		}
		return false
	}
}

// isTensorFrame checks if data represents a tensor frame
func (ts *TensorSerializer) isTensorFrame(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	// Check for tensor frame marker (version=1, marker=255)
	return data[0] == FrameVersion1 && data[1] == 255
}

// serializeTensorMessage serializes a tensor message
func (ts *TensorSerializer) serializeTensorMessage(msg *Message) ([]byte, error) {
	// Extract tensor data and metadata from message
	tensorData, metadata, err := ts.extractTensorFromMessage(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to extract tensor data: %w", err)
	}

	// Serialize tensor data with optimizations
	serializedTensor, err := ts.SerializeTensor(tensorData, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize tensor: %w", err)
	}

	// Create message wrapper with tensor frame
	msgWrapper := map[string]interface{}{
		"message": map[string]interface{}{
			"id":          msg.ID,
			"type":        string(msg.Type),
			"protocol":    string(msg.Protocol),
			"source":      msg.Source.String(),
			"destination": msg.Destination.String(),
			"timestamp":   msg.Timestamp,
			"headers":     msg.Headers,
		},
		"tensor_frame": serializedTensor,
	}

	// Serialize message wrapper as JSON
	msgBytes, err := json.Marshal(msgWrapper)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message wrapper: %w", err)
	}

	return msgBytes, nil
}

// deserializeTensorMessage deserializes a tensor message
func (ts *TensorSerializer) deserializeTensorMessage(data []byte) (*Message, error) {
	// First, try to parse as complete tensor frame
	if ts.isTensorFrame(data) {
		tensorData, metadata, _, err := ts.parseTensorFrame(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse tensor frame: %w", err)
		}

		// Encode tensor data back as payload
		payloadData := map[string]interface{}{
			"tensor_data": tensorData,
			"metadata":    metadata,
		}
		payloadBytes, err := json.Marshal(payloadData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tensor payload: %w", err)
		}

		// Create message with tensor data
		msg := &Message{
			Type:    MessageTypeTensorStream,
			Payload: payloadBytes,
			Headers: map[string]string{
				"Content-Type": "application/tensor",
			},
		}
		return msg, nil
	}

	// Parse as JSON message wrapper with embedded tensor frame
	var msgWrapper map[string]interface{}
	if err := json.Unmarshal(data, &msgWrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message wrapper: %w", err)
	}

	// Extract message metadata
	msgData, ok := msgWrapper["message"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid message wrapper format")
	}

	// Reconstruct message
	msg := &Message{
		Headers: make(map[string]string),
	}

	if id, ok := msgData["id"].(string); ok {
		msg.ID = id
	}
	if msgType, ok := msgData["type"].(string); ok {
		msg.Type = MessageType(msgType)
	}
	if headers, ok := msgData["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if strVal, ok := v.(string); ok {
				msg.Headers[k] = strVal
			}
		}
	}

	// Extract and deserialize tensor frame if present
	if tensorFrameData, hasTensorFrame := msgWrapper["tensor_frame"]; hasTensorFrame {
		var frameBytes []byte

		// Handle different possible formats for tensor_frame data
		switch v := tensorFrameData.(type) {
		case []byte:
			frameBytes = v
		case string:
			// Base64 encoded
			var err error
			frameBytes, err = json.Marshal(v) // This will at least give us something
			if err != nil {
				return nil, fmt.Errorf("failed to handle tensor frame data: %w", err)
			}
		default:
			// JSON serialize whatever it is
			var err error
			frameBytes, err = json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to serialize tensor frame data: %w", err)
			}
		}

		tensorData, metadata, _, err := ts.parseTensorFrame(frameBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse embedded tensor frame: %w", err)
		}

		// Encode tensor data as payload
		payloadData := map[string]interface{}{
			"tensor_data": tensorData,
			"metadata":    metadata,
		}
		payloadBytes, err := json.Marshal(payloadData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tensor payload: %w", err)
		}
		msg.Payload = payloadBytes
		msg.Headers["Content-Type"] = "application/tensor"
	}

	return msg, nil
}

// extractTensorFromMessage extracts tensor data and metadata from a message
func (ts *TensorSerializer) extractTensorFromMessage(msg *Message) ([]byte, *TensorMetadata, error) {
	// Check if payload contains tensor data (JSON format)
	if len(msg.Payload) > 0 {
		var payload map[string]interface{}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			// Try to extract tensor data from JSON payload
			var tensorData []byte
			if data, ok := payload["tensor_data"]; ok {
				if bytes, ok := data.([]byte); ok {
					tensorData = bytes
				} else if str, ok := data.(string); ok {
					// Handle base64 or string encoded data
					tensorData = []byte(str)
				}
			} else if data, ok := payload["activation_data"]; ok {
				if bytes, ok := data.([]byte); ok {
					tensorData = bytes
				} else if str, ok := data.(string); ok {
					tensorData = []byte(str)
				}
			}

			if tensorData != nil {
				// Try to extract metadata
				var metadata *TensorMetadata
				if metaData, ok := payload["metadata"]; ok {
					if metaBytes, err := json.Marshal(metaData); err == nil {
						var meta TensorMetadata
						if json.Unmarshal(metaBytes, &meta) == nil {
							metadata = &meta
						}
					}
				}

				// Create default metadata if none provided
				if metadata == nil {
					metadata = &TensorMetadata{
						DType:  DTypeFloat32, // Default assumption
						Shape:  []int64{int64(len(tensorData) / 4)}, // Assume 1D float32
						Layout: "row_major",
					}
				}

				return tensorData, metadata, nil
			}
		}
	}

	// Fallback: treat entire payload as tensor data
	if len(msg.Payload) > 0 {
		metadata := &TensorMetadata{
			DType:  DTypeFloat32,
			Shape:  []int64{int64(len(msg.Payload) / 4)},
			Layout: "row_major",
		}
		return msg.Payload, metadata, nil
	}

	return nil, nil, fmt.Errorf("no tensor data found in message")
}

// Close closes the tensor serializer and releases resources
func (ts *TensorSerializer) Close() error {
	if ts.zstdEncoder != nil {
		ts.zstdEncoder.Close()
	}
	if ts.zstdDecoder != nil {
		ts.zstdDecoder.Close()
	}
	return nil
}

// NewSerializerFactory creates a new serializer factory
func NewSerializerFactory(format SerializationFormat, enableCompression bool) *SerializerFactory {
	return &SerializerFactory{
		defaultFormat:     format,
		enableCompression: enableCompression,
		tensorOptimized:   false,
	}
}

// NewTensorOptimizedSerializerFactory creates a tensor-optimized serializer factory
func NewTensorOptimizedSerializerFactory(enableCompression bool) *SerializerFactory {
	return &SerializerFactory{
		defaultFormat:     FormatTensor,
		enableCompression: enableCompression,
		tensorOptimized:   true,
	}
}

// CreateSerializer creates a serializer of the specified format
func (sf *SerializerFactory) CreateSerializer(format SerializationFormat) MessageSerializer {
	switch format {
	case FormatJSON:
		return NewJSONSerializer(sf.enableCompression)
	case FormatBinary:
		return NewBinarySerializer(sf.enableCompression)
	case FormatTensor:
		ts, err := NewTensorSerializer(sf.enableCompression)
		if err != nil {
			// Fallback to JSON if tensor serializer fails
			return NewJSONSerializer(sf.enableCompression)
		}
		return ts
	default:
		return NewJSONSerializer(sf.enableCompression)
	}
}

// CreateTensorSerializer creates a tensor-optimized serializer
func (sf *SerializerFactory) CreateTensorSerializer() (*TensorSerializer, error) {
	return NewTensorSerializer(sf.enableCompression)
}

// GetDefaultSerializer returns the default serializer
func (sf *SerializerFactory) GetDefaultSerializer() MessageSerializer {
	return sf.CreateSerializer(sf.defaultFormat)
}

// AddFragment adds a fragment to the reassembly manager and returns complete data if ready
func (fm *FragmentManager) AddFragment(frame *MessageFrame) ([]byte, error) {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	fragmentID := frame.FragmentID
	if fragmentID == "" {
		return nil, fmt.Errorf("missing fragment ID")
	}

	// Get or create partial message
	partial, exists := fm.fragments[fragmentID]
	if !exists {
		partial = &PartialMessage{
			Fragments:      make(map[uint16]*MessageFrame),
			TotalFragments: frame.TotalFragments,
			LastActivity:   time.Now(),
			Complete:       false,
		}
		fm.fragments[fragmentID] = partial
		fm.timeouts[fragmentID] = time.Now().Add(FragmentTimeout)
	}

	// Validate fragment
	if frame.FragmentIndex >= frame.TotalFragments {
		return nil, fmt.Errorf("invalid fragment index %d >= %d", frame.FragmentIndex, frame.TotalFragments)
	}
	if frame.TotalFragments != partial.TotalFragments {
		return nil, fmt.Errorf("fragment count mismatch: expected %d, got %d", partial.TotalFragments, frame.TotalFragments)
	}

	// Store fragment
	partial.Fragments[frame.FragmentIndex] = frame
	partial.LastActivity = time.Now()

	// Check if message is complete
	if len(partial.Fragments) == int(partial.TotalFragments) {
		// Reassemble complete message
		completeData, err := fm.assembleFragments(partial)
		if err != nil {
			// Clean up failed assembly
			delete(fm.fragments, fragmentID)
			delete(fm.timeouts, fragmentID)
			return nil, fmt.Errorf("failed to assemble fragments: %w", err)
		}

		// Clean up completed message
		delete(fm.fragments, fragmentID)
		delete(fm.timeouts, fragmentID)
		partial.Complete = true

		return completeData, nil
	}

	return nil, nil // Message not yet complete
}

// assembleFragments assembles fragments into complete message data
func (fm *FragmentManager) assembleFragments(partial *PartialMessage) ([]byte, error) {
	// Calculate total size
	totalSize := 0
	for i := uint16(0); i < partial.TotalFragments; i++ {
		fragment, exists := partial.Fragments[i]
		if !exists {
			return nil, fmt.Errorf("missing fragment %d", i)
		}
		totalSize += len(fragment.Data)
	}

	// Assemble data in order
	completeData := make([]byte, 0, totalSize)
	for i := uint16(0); i < partial.TotalFragments; i++ {
		fragment := partial.Fragments[i]
		completeData = append(completeData, fragment.Data...)
	}

	return completeData, nil
}

// cleanupRoutine periodically removes expired fragments
func (fm *FragmentManager) cleanupRoutine() {
	for {
		select {
		case <-fm.done:
			return
		case <-fm.cleanupTicker.C:
			fm.cleanupExpiredFragments()
		}
	}
}

// cleanupExpiredFragments removes fragments that have exceeded their timeout
func (fm *FragmentManager) cleanupExpiredFragments() {
	fm.mutex.Lock()
	defer fm.mutex.Unlock()

	now := time.Now()
	for fragmentID, expiry := range fm.timeouts {
		if now.After(expiry) {
			delete(fm.fragments, fragmentID)
			delete(fm.timeouts, fragmentID)
		}
	}
}

// Close shuts down the fragment manager
func (fm *FragmentManager) Close() error {
	close(fm.done)
	fm.cleanupTicker.Stop()
	return nil
}
