package messaging

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
)

// JSONSerializer implements JSON-based message serialization
type JSONSerializer struct {
	enableCompression bool
	compressionLevel  int
}

// BinarySerializer implements binary message serialization
type BinarySerializer struct {
	enableCompression bool
	compressionLevel  int
}

// MessageFrame represents a framed message for transmission
type MessageFrame struct {
	Version     uint8  `json:"version"`
	MessageType uint8  `json:"message_type"`
	Flags       uint8  `json:"flags"`
	Length      uint32 `json:"length"`
	Checksum    uint32 `json:"checksum"`
	Data        []byte `json:"data"`
}

// Serialization flags
const (
	FlagCompressed = 1 << iota
	FlagEncrypted
	FlagFragmented
)

// Message frame version
const (
	FrameVersion1 = 1
)

// NewJSONSerializer creates a new JSON serializer
func NewJSONSerializer(enableCompression bool) *JSONSerializer {
	return &JSONSerializer{
		enableCompression: enableCompression,
		compressionLevel:  gzip.DefaultCompression,
	}
}

// Serialize serializes a message to bytes
func (js *JSONSerializer) Serialize(msg *Message) ([]byte, error) {
	// Marshal message to JSON
	data, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Compress if enabled and beneficial
	if js.enableCompression && len(data) > 1024 {
		compressed, err := js.compress(data)
		if err != nil {
			return nil, fmt.Errorf("failed to compress message: %w", err)
		}

		// Use compressed data if it's smaller
		if len(compressed) < len(data) {
			data = compressed

			// Create frame with compression flag
			frame := &MessageFrame{
				Version:     FrameVersion1,
				MessageType: uint8(getMessageTypeCode(msg.Type)),
				Flags:       FlagCompressed,
				Length:      uint32(len(data)),
				Checksum:    calculateChecksum(data),
				Data:        data,
			}

			return js.serializeFrame(frame)
		}
	}

	// Create uncompressed frame
	frame := &MessageFrame{
		Version:     FrameVersion1,
		MessageType: uint8(getMessageTypeCode(msg.Type)),
		Flags:       0,
		Length:      uint32(len(data)),
		Checksum:    calculateChecksum(data),
		Data:        data,
	}

	return js.serializeFrame(frame)
}

// Deserialize deserializes bytes to a message
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

	// Decompress if needed
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
	if len(data) < 10 { // Minimum frame size
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

// SerializerFactory creates serializers based on configuration
type SerializerFactory struct {
	defaultFormat     SerializationFormat
	enableCompression bool
}

type SerializationFormat string

const (
	FormatJSON   SerializationFormat = "json"
	FormatBinary SerializationFormat = "binary"
)

// NewSerializerFactory creates a new serializer factory
func NewSerializerFactory(format SerializationFormat, enableCompression bool) *SerializerFactory {
	return &SerializerFactory{
		defaultFormat:     format,
		enableCompression: enableCompression,
	}
}

// CreateSerializer creates a serializer of the specified format
func (sf *SerializerFactory) CreateSerializer(format SerializationFormat) MessageSerializer {
	switch format {
	case FormatJSON:
		return NewJSONSerializer(sf.enableCompression)
	case FormatBinary:
		return NewBinarySerializer(sf.enableCompression)
	default:
		return NewJSONSerializer(sf.enableCompression)
	}
}

// GetDefaultSerializer returns the default serializer
func (sf *SerializerFactory) GetDefaultSerializer() MessageSerializer {
	return sf.CreateSerializer(sf.defaultFormat)
}
