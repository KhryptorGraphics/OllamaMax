package protocols

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// CompressionType defines the type of compression used for tensor data
type CompressionType uint8

const (
	CompressionNone CompressionType = iota
	CompressionGzip
	CompressionZstd
	CompressionQuantized
	CompressionSparse
	CompressionTopK
)

// TensorDType represents the data type of tensor elements
type TensorDType uint8

const (
	DTypeFloat32 TensorDType = iota
	DTypeFloat16
	DTypeInt8
	DTypeBool
)

// TensorMetadata contains information about tensor structure and compression
type TensorMetadata struct {
	Shape           []int64         `json:"shape"`
	DType           TensorDType     `json:"dtype"`
	CompressionType CompressionType `json:"compression_type"`
	OriginalSize    int64           `json:"original_size"`
	CompressedSize  int64           `json:"compressed_size"`
	Sparsity        float32         `json:"sparsity,omitempty"`
	TopKRatio       float32         `json:"topk_ratio,omitempty"`
}

// TensorCompressor handles compression and decompression of tensor data
type TensorCompressor struct {
	gzipPool *sync.Pool
	zstdPool *sync.Pool
}

// NewTensorCompressor creates a new tensor compressor with pooled resources
func NewTensorCompressor() *TensorCompressor {
	return &TensorCompressor{
		gzipPool: &sync.Pool{
			New: func() interface{} {
				return gzip.NewWriter(nil)
			},
		},
		zstdPool: &sync.Pool{
			New: func() interface{} {
				encoder, _ := zstd.NewWriter(nil)
				return encoder
			},
		},
	}
}

// CompressActivation compresses tensor data using the optimal compression strategy
func (tc *TensorCompressor) CompressActivation(data []byte, metadata *TensorMetadata) ([]byte, error) {
	switch metadata.CompressionType {
	case CompressionNone:
		return data, nil
	case CompressionGzip:
		return tc.compressGzip(data)
	case CompressionZstd:
		return tc.compressZstd(data)
	case CompressionQuantized:
		return tc.compressQuantized(data, metadata)
	case CompressionSparse:
		return tc.compressSparse(data, metadata)
	case CompressionTopK:
		return tc.compressTopK(data, metadata)
	default:
		return nil, fmt.Errorf("unsupported compression type: %v", metadata.CompressionType)
	}
}

// DecompressActivation decompresses tensor data
func (tc *TensorCompressor) DecompressActivation(data []byte, metadata *TensorMetadata) ([]byte, error) {
	switch metadata.CompressionType {
	case CompressionNone:
		return data, nil
	case CompressionGzip:
		return tc.decompressGzip(data)
	case CompressionZstd:
		return tc.decompressZstd(data)
	case CompressionQuantized:
		return tc.decompressQuantized(data, metadata)
	case CompressionSparse:
		return tc.decompressSparse(data, metadata)
	case CompressionTopK:
		return tc.decompressTopK(data, metadata)
	default:
		return nil, fmt.Errorf("unsupported compression type: %v", metadata.CompressionType)
	}
}

// compressGzip compresses data using gzip
func (tc *TensorCompressor) compressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := tc.gzipPool.Get().(*gzip.Writer)
	defer tc.gzipPool.Put(writer)
	
	writer.Reset(&buf)
	if _, err := writer.Write(data); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// decompressGzip decompresses gzip data
func (tc *TensorCompressor) decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// compressZstd compresses data using zstd
func (tc *TensorCompressor) compressZstd(data []byte) ([]byte, error) {
	encoder := tc.zstdPool.Get().(*zstd.Encoder)
	defer tc.zstdPool.Put(encoder)
	
	var buf bytes.Buffer
	encoder.Reset(&buf)
	if _, err := encoder.Write(data); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// decompressZstd decompresses zstd data
func (tc *TensorCompressor) decompressZstd(data []byte) ([]byte, error) {
	decoder, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer decoder.Close()
	
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(decoder); err != nil {
		return nil, err
	}
	
	return buf.Bytes(), nil
}

// compressQuantized applies quantization compression
func (tc *TensorCompressor) compressQuantized(data []byte, metadata *TensorMetadata) ([]byte, error) {
	if metadata.DType != DTypeFloat32 {
		return nil, fmt.Errorf("quantization only supported for float32 tensors")
	}

	// Convert float32 to int8 quantization
	floats := make([]float32, len(data)/4)
	buf := bytes.NewReader(data)
	if err := binary.Read(buf, binary.LittleEndian, floats); err != nil {
		return nil, err
	}

	if len(floats) == 0 {
		return nil, fmt.Errorf("no data to quantize")
	}

	// Find min/max for quantization
	min, max := floats[0], floats[0]
	for _, f := range floats {
		if f < min {
			min = f
		}
		if f > max {
			max = f
		}
	}

	// Handle constant tensor edge case (when max == min)
	var result bytes.Buffer
	binary.Write(&result, binary.LittleEndian, min)
	binary.Write(&result, binary.LittleEndian, max)
	binary.Write(&result, binary.LittleEndian, int32(len(floats))) // Store original length

	if max == min {
		// Constant tensor: encode scale=0 as marker and store min value once
		binary.Write(&result, binary.LittleEndian, float32(0)) // scale = 0 indicates constant
		// No quantized array needed for constant tensors
		return result.Bytes(), nil
	}

	// Regular quantization
	scale := (max - min) / 255.0
	binary.Write(&result, binary.LittleEndian, scale)

	quantized := make([]int8, len(floats))
	for i, f := range floats {
		quantized[i] = int8((f - min) / scale)
	}

	binary.Write(&result, binary.LittleEndian, quantized)

	return result.Bytes(), nil
}

// decompressQuantized decompresses quantized data
func (tc *TensorCompressor) decompressQuantized(data []byte, metadata *TensorMetadata) ([]byte, error) {
	buf := bytes.NewReader(data)

	var min, max float32
	var originalLength int32
	var scale float32

	if err := binary.Read(buf, binary.LittleEndian, &min); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &max); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &originalLength); err != nil {
		return nil, err
	}
	if err := binary.Read(buf, binary.LittleEndian, &scale); err != nil {
		return nil, err
	}

	// Check for constant tensor (scale == 0)
	if scale == 0 {
		// Reconstruct constant tensor filled with min value
		floats := make([]float32, originalLength)
		for i := range floats {
			floats[i] = min
		}

		// Serialize back to bytes
		var result bytes.Buffer
		if err := binary.Write(&result, binary.LittleEndian, floats); err != nil {
			return nil, err
		}
		return result.Bytes(), nil
	}

	// Regular dequantization
	quantizedSize := len(data) - 16 // 4 bytes each for min, max, length, scale
	quantized := make([]int8, quantizedSize)
	if err := binary.Read(buf, binary.LittleEndian, quantized); err != nil {
		return nil, err
	}

	// Dequantize
	floats := make([]float32, len(quantized))
	for i, q := range quantized {
		floats[i] = min + float32(q)*scale
	}

	// Serialize back to bytes
	var result bytes.Buffer
	if err := binary.Write(&result, binary.LittleEndian, floats); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

// compressSparse compresses sparse tensors using coordinate format
func (tc *TensorCompressor) compressSparse(data []byte, metadata *TensorMetadata) ([]byte, error) {
	// Implementation for sparse tensor compression
	// This is a simplified version - real implementation would handle various sparse formats
	return tc.compressZstd(data) // Fallback to zstd for now
}

// decompressSparse decompresses sparse tensors
func (tc *TensorCompressor) decompressSparse(data []byte, metadata *TensorMetadata) ([]byte, error) {
	// Implementation for sparse tensor decompression
	return tc.decompressZstd(data) // Fallback to zstd for now
}

// compressTopK applies top-k sparsification
func (tc *TensorCompressor) compressTopK(data []byte, metadata *TensorMetadata) ([]byte, error) {
	// Implementation for top-k compression
	return tc.compressZstd(data) // Fallback to zstd for now
}

// decompressTopK decompresses top-k compressed data
func (tc *TensorCompressor) decompressTopK(data []byte, metadata *TensorMetadata) ([]byte, error) {
	// Implementation for top-k decompression
	return tc.decompressZstd(data) // Fallback to zstd for now
}

// ActivationOptimizer analyzes tensor patterns to select optimal compression
type ActivationOptimizer struct {
	compressor *TensorCompressor
}

// NewActivationOptimizer creates a new activation optimizer
func NewActivationOptimizer() *ActivationOptimizer {
	return &ActivationOptimizer{
		compressor: NewTensorCompressor(),
	}
}

// OptimizeCompression selects the best compression strategy for given tensor data
func (ao *ActivationOptimizer) OptimizeCompression(data []byte, shape []int64, dtype TensorDType) (*TensorMetadata, error) {
	metadata := &TensorMetadata{
		Shape:        shape,
		DType:        dtype,
		OriginalSize: int64(len(data)),
	}
	
	// Analyze tensor characteristics
	sparsity := ao.calculateSparsity(data, dtype)
	metadata.Sparsity = sparsity
	
	// Select compression strategy based on characteristics
	if sparsity > 0.8 {
		metadata.CompressionType = CompressionSparse
	} else if dtype == DTypeFloat32 && len(data) > 1024*1024 { // Large float32 tensors
		metadata.CompressionType = CompressionQuantized
	} else {
		metadata.CompressionType = CompressionZstd
	}
	
	return metadata, nil
}

// calculateSparsity calculates the sparsity ratio of tensor data
func (ao *ActivationOptimizer) calculateSparsity(data []byte, dtype TensorDType) float32 {
	switch dtype {
	case DTypeFloat32:
		// Validate length alignment
		if len(data)%4 != 0 {
			// Truncate to aligned length
			alignedLen := (len(data) / 4) * 4
			if alignedLen == 0 {
				return 0.0 // Conservative default for invalid data
			}
			data = data[:alignedLen]
		}

		floats := make([]float32, len(data)/4)
		buf := bytes.NewReader(data)

		// Handle binary.Read error
		if err := binary.Read(buf, binary.LittleEndian, floats); err != nil {
			return 0.0 // Conservative default on read error
		}

		zeros := 0
		for _, f := range floats {
			if math.Abs(float64(f)) < 1e-6 {
				zeros++
			}
		}
		return float32(zeros) / float32(len(floats))
	default:
		return 0.0 // Conservative estimate for other types
	}
}
