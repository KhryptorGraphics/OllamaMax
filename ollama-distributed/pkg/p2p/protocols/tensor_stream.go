package protocols

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
)

// Tensor stream message types
const (
	MsgTypeActivationStart    = "activation_start"
	MsgTypeActivationChunk    = "activation_chunk"
	MsgTypeActivationComplete = "activation_complete"
	MsgTypeActivationError    = "activation_error"
	MsgTypeActivationAck      = "activation_ack"
)

// ActivationHeader contains metadata for activation streaming
type ActivationHeader struct {
	InferenceID string          `json:"inference_id"`
	PartitionID string          `json:"partition_id"`
	TensorShape []int64         `json:"tensor_shape"`
	DType       TensorDType     `json:"dtype"`
	Compression CompressionType `json:"compression"`
	TotalSize   int64           `json:"total_size"`
	ChunkCount  int32           `json:"chunk_count"`
	SequenceNum int32           `json:"sequence_num"`
	Timestamp   time.Time       `json:"timestamp"`
	SourcePeer  peer.ID         `json:"source_peer"`
	TargetPeer  peer.ID         `json:"target_peer"`
	Priority    int32           `json:"priority"`
}

// ActivationChunk represents a chunk of activation data
type ActivationChunk struct {
	Header     ActivationHeader `json:"header"`
	ChunkIndex int32            `json:"chunk_index"`
	ChunkSize  int32            `json:"chunk_size"`
	Data       []byte           `json:"data"`
	Checksum   string           `json:"checksum"`
}

// ActivationStreamMessage represents messages in the tensor stream protocol
type ActivationStreamMessage struct {
	Type      string            `json:"type"`
	Header    *ActivationHeader `json:"header,omitempty"`
	Chunk     *ActivationChunk  `json:"chunk,omitempty"`
	Error     string            `json:"error,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// Protocol ID constant
const ProtocolTensorStream protocol.ID = "/ollama-distributed/tensor-stream/1.0.0"

// Binary framing constants
const (
	BinaryFrameVersion    = 1
	BinaryFrameHeaderSize = 16 // Version(1) + Type(1) + Flags(1) + Reserved(1) + Length(4) + Checksum(4) + Timestamp(4)

	// Frame flags
	FrameFlagCompressed = 1 << 0
	FrameFlagEncrypted  = 1 << 1
	FrameFlagPriority   = 1 << 2
)

// BinaryFrameHeader represents the binary frame header
type BinaryFrameHeader struct {
	Version   uint8  // Frame format version
	Type      uint8  // Message type (mapped from string constants)
	Flags     uint8  // Frame flags
	Reserved  uint8  // Reserved for future use
	Length    uint32 // Payload length
	Checksum  uint32 // CRC32 checksum of payload
	Timestamp uint32 // Unix timestamp (seconds)
}

// BinaryActivationChunk represents a chunk of activation data in pure binary format
type BinaryActivationChunk struct {
	InferenceIDLen uint8     // Length of inference ID (up to 255 chars)
	InferenceID    []byte    // Inference ID as raw bytes
	PartitionIDLen uint8     // Length of partition ID (up to 255 chars)
	PartitionID    []byte    // Partition ID as raw bytes
	ChunkIndex     uint32    // Chunk index
	ChunkSize      uint32    // Size of data payload
	SequenceNum    uint32    // Sequence number
	Timestamp      uint64    // Unix timestamp (nanoseconds)
	DataChecksum   uint32    // CRC32 checksum of data payload
	Data           []byte    // Raw binary tensor data
}

// TensorStreamProtocol handles tensor streaming between nodes
type TensorStreamProtocol struct {
	host          host.Host
	activeStreams map[string]*ActiveStream
	streamsMutex  sync.RWMutex
	chunkSize     int32
	maxConcurrent int
	compressor    *TensorCompressor
	optimizer     *ActivationOptimizer

	// Response registry for completion notifications
	responseRegistry map[string]chan *ActivationData
	registryMutex    sync.RWMutex

	// Reference to pipeline coordinator for output publishing
	pipelineCoord    *PipelineCoordinator
}

// ActiveStream tracks an ongoing tensor stream
type ActiveStream struct {
	ID            string
	Header        ActivationHeader
	Chunks        map[int32]*ActivationChunk
	ReceivedCount int32
	TotalChunks   int32
	StartTime     time.Time
	LastActivity  time.Time
	Complete      bool
	Error         error
	mutex         sync.RWMutex
}

// NewTensorStreamProtocol creates a new tensor stream protocol handler
func NewTensorStreamProtocol(h host.Host) *TensorStreamProtocol {
	tsp := &TensorStreamProtocol{
		host:             h,
		activeStreams:    make(map[string]*ActiveStream),
		chunkSize:        1024 * 1024, // 1MB default chunk size
		maxConcurrent:    10,
		compressor:       NewTensorCompressor(),
		optimizer:        NewActivationOptimizer(),
		responseRegistry: make(map[string]chan *ActivationData),
	}

	// Register stream handler
	h.SetStreamHandler(ProtocolTensorStream, tsp.HandleStream)

	return tsp
}

// SetPipelineCoordinator sets the pipeline coordinator reference for publishing outputs
func (tsp *TensorStreamProtocol) SetPipelineCoordinator(pc *PipelineCoordinator) {
	tsp.pipelineCoord = pc
}

// RegisterResponseChannel registers a response channel for a specific stream ID
func (tsp *TensorStreamProtocol) RegisterResponseChannel(streamID string) (chan *ActivationData, func()) {
	tsp.registryMutex.Lock()
	defer tsp.registryMutex.Unlock()

	responseChan := make(chan *ActivationData, 1)
	tsp.responseRegistry[streamID] = responseChan

	unregister := func() {
		tsp.registryMutex.Lock()
		defer tsp.registryMutex.Unlock()
		delete(tsp.responseRegistry, streamID)
		close(responseChan)
	}

	return responseChan, unregister
}

// HandleStream handles incoming tensor stream connections
func (tsp *TensorStreamProtocol) HandleStream(stream network.Stream) {
	defer stream.Close()

	reader := bufio.NewReader(stream)
	writer := bufio.NewWriter(stream)

	for {
		// Read message
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading tensor stream message: %v", err)
			}
			break
		}

		// Use binary framing for all messages (true binary chunk framing without JSON/base64)
		msg, err := decodeBinaryFrame(line[:len(line)-1])
		if err != nil {
			log.Printf("Error decoding binary tensor stream message: %v", err)
			continue
		}

		// Handle message based on type
		response := tsp.handleMessage(msg, stream.Conn().RemotePeer())
		if response != nil {
			// Use binary framing for response
			responseData, err := encodeBinaryFrame(response)
			if err != nil {
				log.Printf("Error encoding binary response: %v", err)
				// Fall back to JSON
				responseData, _ = json.Marshal(response)
				writer.Write(append(responseData, '\n'))
			} else {
				writer.Write(append(responseData, '\n'))
			}
			writer.Flush()
		}
	}
}

// handleMessage processes different types of tensor stream messages
func (tsp *TensorStreamProtocol) handleMessage(msg *ActivationStreamMessage, remotePeer peer.ID) *ActivationStreamMessage {
	switch msg.Type {
	case MsgTypeActivationStart:
		return tsp.handleActivationStart(msg, remotePeer)
	case MsgTypeActivationChunk:
		return tsp.handleActivationChunk(msg, remotePeer)
	case MsgTypeActivationComplete:
		return tsp.handleActivationComplete(msg, remotePeer)
	case MsgTypeActivationError:
		return tsp.handleActivationError(msg, remotePeer)
	case MsgTypeActivationAck:
		return tsp.handleActivationAck(msg, remotePeer)
	default:
		log.Printf("Unknown tensor stream message type: %s", msg.Type)
		return &ActivationStreamMessage{
			Type:      MsgTypeActivationError,
			Error:     fmt.Sprintf("Unknown message type: %s", msg.Type),
			Timestamp: time.Now(),
		}
	}
}

// handleActivationStart initializes a new activation stream
func (tsp *TensorStreamProtocol) handleActivationStart(msg *ActivationStreamMessage, remotePeer peer.ID) *ActivationStreamMessage {
	if msg.Header == nil {
		return &ActivationStreamMessage{
			Type:      MsgTypeActivationError,
			Error:     "Missing activation header",
			Timestamp: time.Now(),
		}
	}

	streamID := fmt.Sprintf("%s-%s-%d", msg.Header.InferenceID, msg.Header.PartitionID, msg.Header.SequenceNum)

	tsp.streamsMutex.Lock()
	defer tsp.streamsMutex.Unlock()

	// Check if stream already exists
	if _, exists := tsp.activeStreams[streamID]; exists {
		return &ActivationStreamMessage{
			Type:      MsgTypeActivationError,
			Error:     "Stream already exists",
			Timestamp: time.Now(),
		}
	}

	// Create new active stream
	stream := &ActiveStream{
		ID:            streamID,
		Header:        *msg.Header,
		Chunks:        make(map[int32]*ActivationChunk),
		ReceivedCount: 0,
		TotalChunks:   msg.Header.ChunkCount,
		StartTime:     time.Now(),
		LastActivity:  time.Now(),
		Complete:      false,
	}

	tsp.activeStreams[streamID] = stream

	log.Printf("Started tensor stream: %s from peer %s", streamID, remotePeer)

	return &ActivationStreamMessage{
		Type:      MsgTypeActivationAck,
		Header:    msg.Header,
		Timestamp: time.Now(),
	}
}

// handleActivationChunk processes incoming activation chunks
func (tsp *TensorStreamProtocol) handleActivationChunk(msg *ActivationStreamMessage, remotePeer peer.ID) *ActivationStreamMessage {
	if msg.Chunk == nil {
		return &ActivationStreamMessage{
			Type:      MsgTypeActivationError,
			Error:     "Missing chunk data",
			Timestamp: time.Now(),
		}
	}

	streamID := fmt.Sprintf("%s-%s-%d", msg.Chunk.Header.InferenceID, msg.Chunk.Header.PartitionID, msg.Chunk.Header.SequenceNum)

	tsp.streamsMutex.Lock()
	defer tsp.streamsMutex.Unlock()

	stream, exists := tsp.activeStreams[streamID]
	if !exists {
		return &ActivationStreamMessage{
			Type:      MsgTypeActivationError,
			Error:     "Stream not found",
			Timestamp: time.Now(),
		}
	}

	stream.mutex.Lock()
	defer stream.mutex.Unlock()

	// Store chunk
	stream.Chunks[msg.Chunk.ChunkIndex] = msg.Chunk
	stream.ReceivedCount++
	stream.LastActivity = time.Now()

	log.Printf("Received chunk %d/%d for stream %s", msg.Chunk.ChunkIndex+1, stream.TotalChunks, streamID)

	// Check if all chunks received
	if stream.ReceivedCount == stream.TotalChunks {
		stream.Complete = true
		log.Printf("Stream %s completed", streamID)
	}

	return &ActivationStreamMessage{
		Type:      MsgTypeActivationAck,
		Header:    &msg.Chunk.Header,
		Timestamp: time.Now(),
	}
}

// handleActivationComplete marks a stream as complete and publishes results
func (tsp *TensorStreamProtocol) handleActivationComplete(msg *ActivationStreamMessage, remotePeer peer.ID) *ActivationStreamMessage {
	if msg.Header == nil {
		return &ActivationStreamMessage{
			Type:      MsgTypeActivationError,
			Error:     "Missing activation header",
			Timestamp: time.Now(),
		}
	}

	streamID := fmt.Sprintf("%s-%s-%d", msg.Header.InferenceID, msg.Header.PartitionID, msg.Header.SequenceNum)

	tsp.streamsMutex.Lock()
	defer tsp.streamsMutex.Unlock()

	if stream, exists := tsp.activeStreams[streamID]; exists {
		stream.mutex.Lock()
		stream.Complete = true
		stream.mutex.Unlock()

		log.Printf("Stream %s marked as complete, assembling chunks", streamID)

		// Assemble received chunks into complete tensor data
		assembledData, err := tsp.assembleStreamChunks(stream)
		if err != nil {
			log.Printf("Failed to assemble chunks for stream %s: %v", streamID, err)
			return &ActivationStreamMessage{
				Type:      MsgTypeActivationError,
				Error:     fmt.Sprintf("Failed to assemble chunks: %v", err),
				Timestamp: time.Now(),
			}
		}

		// Create ActivationData from assembled tensor
		activationData := &ActivationData{
			InferenceID: msg.Header.InferenceID,
			StageID:     msg.Header.PartitionID,
			Data:        assembledData,
			Metadata: map[string]interface{}{
				"tensor_shape":      msg.Header.TensorShape,
				"dtype":             msg.Header.DType,
				"compression":       msg.Header.Compression,
				"original_size":     msg.Header.TotalSize,
				"source_peer":       msg.Header.SourcePeer.String(),
				"target_peer":       msg.Header.TargetPeer.String(),
				"sequence_num":      msg.Header.SequenceNum,
			},
			Timestamp: time.Now(),
		}

		// Publish to response channel if registered
		tsp.publishToResponseChannel(streamID, activationData)

		// Publish to pipeline coordinator stage buffer if available
		tsp.publishToPipelineCoordinator(activationData)
	}

	return &ActivationStreamMessage{
		Type:      MsgTypeActivationAck,
		Header:    msg.Header,
		Timestamp: time.Now(),
	}
}

// handleActivationError handles error messages
func (tsp *TensorStreamProtocol) handleActivationError(msg *ActivationStreamMessage, remotePeer peer.ID) *ActivationStreamMessage {
	log.Printf("Received activation error from %s: %s", remotePeer, msg.Error)
	return nil // No response needed for error messages
}

// handleActivationAck handles acknowledgment messages
func (tsp *TensorStreamProtocol) handleActivationAck(msg *ActivationStreamMessage, remotePeer peer.ID) *ActivationStreamMessage {
	// Process acknowledgment - could be used for flow control
	return nil // No response needed for acks
}

// GetActiveStream retrieves an active stream by ID
func (tsp *TensorStreamProtocol) GetActiveStream(streamID string) (*ActiveStream, bool) {
	tsp.streamsMutex.RLock()
	defer tsp.streamsMutex.RUnlock()

	stream, exists := tsp.activeStreams[streamID]
	return stream, exists
}

// CleanupStream removes a completed or failed stream
func (tsp *TensorStreamProtocol) CleanupStream(streamID string) {
	tsp.streamsMutex.Lock()
	defer tsp.streamsMutex.Unlock()

	delete(tsp.activeStreams, streamID)
	log.Printf("Cleaned up stream: %s", streamID)
}

// GetStreamStats returns statistics about active streams
func (tsp *TensorStreamProtocol) GetStreamStats() map[string]interface{} {
	tsp.streamsMutex.RLock()
	defer tsp.streamsMutex.RUnlock()

	stats := map[string]interface{}{
		"active_streams": len(tsp.activeStreams),
		"chunk_size":     tsp.chunkSize,
		"max_concurrent": tsp.maxConcurrent,
	}

	return stats
}

// SetChunkSize configures the chunk size for tensor streaming
func (tsp *TensorStreamProtocol) SetChunkSize(size int32) {
	tsp.chunkSize = size
}

// SetMaxConcurrent configures the maximum concurrent streams
func (tsp *TensorStreamProtocol) SetMaxConcurrent(max int) {
	tsp.maxConcurrent = max
}

// SendMessage sends a message to a peer using libp2p streams
func (tsp *TensorStreamProtocol) SendMessage(ctx context.Context, peerID peer.ID, msg *ActivationStreamMessage) error {
	// Create a new stream to the peer
	s, err := tsp.host.NewStream(ctx, peerID, ProtocolTensorStream)
	if err != nil {
		return fmt.Errorf("failed to create stream to peer %s: %w", peerID, err)
	}
	defer s.Close()

	// Create buffered writer
	bw := bufio.NewWriter(s)

	// Encode message using binary framing
	b, err := encodeBinaryFrame(msg)
	if err != nil {
		return fmt.Errorf("failed to encode binary frame: %w", err)
	}

	// Write binary frame with newline delimiter (for stream boundaries)
	_, err = bw.Write(append(b, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	// Flush the buffer
	err = bw.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush message: %w", err)
	}

	log.Printf("Successfully sent message to peer %s: %s", peerID, msg.Type)
	return nil
}

// assembleStreamChunks assembles all received chunks into complete tensor data
func (tsp *TensorStreamProtocol) assembleStreamChunks(stream *ActiveStream) ([]byte, error) {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()

	if !stream.Complete || stream.ReceivedCount != stream.TotalChunks {
		return nil, fmt.Errorf("stream not complete: received %d/%d chunks", stream.ReceivedCount, stream.TotalChunks)
	}

	// Calculate total size
	totalSize := int64(0)
	for _, chunk := range stream.Chunks {
		totalSize += int64(chunk.ChunkSize)
	}

	// Assemble chunks in order
	assembledData := make([]byte, 0, totalSize)
	for i := int32(0); i < stream.TotalChunks; i++ {
		if chunk, exists := stream.Chunks[i]; exists {
			assembledData = append(assembledData, chunk.Data...)
		} else {
			return nil, fmt.Errorf("missing chunk %d", i)
		}
	}

	// Decompress if needed
	if stream.Header.Compression != CompressionNone {
		if tsp.compressor != nil {
			decompressed, err := tsp.compressor.DecompressActivation(assembledData, &TensorMetadata{
				Shape:           stream.Header.TensorShape,
				DType:           TensorDType(stream.Header.DType),
				CompressionType: CompressionType(stream.Header.Compression),
				OriginalSize:    stream.Header.TotalSize,
			})
			if err != nil {
				return nil, fmt.Errorf("decompression failed: %w", err)
			}
			return decompressed, nil
		}
	}

	return assembledData, nil
}

// publishToResponseChannel publishes activation data to registered response channel
func (tsp *TensorStreamProtocol) publishToResponseChannel(streamID string, data *ActivationData) {
	tsp.registryMutex.RLock()
	responseChan, exists := tsp.responseRegistry[streamID]
	tsp.registryMutex.RUnlock()

	if exists {
		select {
		case responseChan <- data:
			log.Printf("Published activation data to response channel for stream %s", streamID)
		default:
			log.Printf("Response channel full for stream %s, skipping", streamID)
		}
	}
}

// publishToPipelineCoordinator publishes activation data to pipeline stage buffer
func (tsp *TensorStreamProtocol) publishToPipelineCoordinator(data *ActivationData) {
	if tsp.pipelineCoord != nil {
		stageID := data.StageID
		if buffer := tsp.pipelineCoord.getStageBuffer(stageID); buffer != nil {
			select {
			case buffer.OutputBuffer <- data:
				log.Printf("Published activation data to stage buffer for stage %s", stageID)
			default:
				log.Printf("Stage buffer full for stage %s, skipping", stageID)
			}
		}
	}
}

// Binary framing implementation

// mapMessageType converts string message types to binary type codes
func mapMessageType(msgType string) uint8 {
	switch msgType {
	case MsgTypeActivationStart:
		return 1
	case MsgTypeActivationChunk:
		return 2
	case MsgTypeActivationComplete:
		return 3
	case MsgTypeActivationError:
		return 4
	case MsgTypeActivationAck:
		return 5
	default:
		return 0 // Unknown
	}
}

// unmapMessageType converts binary type codes back to string message types
func unmapMessageType(typeCode uint8) string {
	switch typeCode {
	case 1:
		return MsgTypeActivationStart
	case 2:
		return MsgTypeActivationChunk
	case 3:
		return MsgTypeActivationComplete
	case 4:
		return MsgTypeActivationError
	case 5:
		return MsgTypeActivationAck
	default:
		return "unknown"
	}
}

// encodeBinaryFrame encodes a message using binary framing
func encodeBinaryFrame(msg *ActivationStreamMessage) ([]byte, error) {
	var payload []byte
	var err error

	// For activation chunks, use pure binary encoding to avoid JSON/base64 overhead
	if msg.Type == MsgTypeActivationChunk && msg.Chunk != nil {
		payload, err = encodeBinaryActivationChunk(msg.Chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to encode binary activation chunk: %w", err)
		}
	} else {
		// For control messages, still use JSON
		payload, err = json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to encode message payload: %w", err)
		}
	}

	// Calculate checksum
	checksum := calculateCRC32(payload)

	// Create frame header
	header := BinaryFrameHeader{
		Version:   BinaryFrameVersion,
		Type:      mapMessageType(msg.Type),
		Flags:     0, // No compression/encryption for now
		Reserved:  0,
		Length:    uint32(len(payload)),
		Checksum:  checksum,
		Timestamp: uint32(time.Now().Unix()),
	}

	// Set binary flag for activation chunks
	if msg.Type == MsgTypeActivationChunk {
		header.Flags |= FrameFlagCompressed // Reuse compression flag to indicate binary mode
	}

	// Encode frame header to binary
	headerBuf := make([]byte, BinaryFrameHeaderSize)
	headerBuf[0] = header.Version
	headerBuf[1] = header.Type
	headerBuf[2] = header.Flags
	headerBuf[3] = header.Reserved
	binary.BigEndian.PutUint32(headerBuf[4:8], header.Length)
	binary.BigEndian.PutUint32(headerBuf[8:12], header.Checksum)
	binary.BigEndian.PutUint32(headerBuf[12:16], header.Timestamp)

	// Combine header and payload
	frame := make([]byte, 0, BinaryFrameHeaderSize+len(payload))
	frame = append(frame, headerBuf...)
	frame = append(frame, payload...)

	return frame, nil
}

// decodeBinaryFrame decodes a binary frame back to a message
func decodeBinaryFrame(frame []byte) (*ActivationStreamMessage, error) {
	if len(frame) < BinaryFrameHeaderSize {
		return nil, fmt.Errorf("frame too short: got %d bytes, need at least %d", len(frame), BinaryFrameHeaderSize)
	}

	// Decode frame header
	header := BinaryFrameHeader{
		Version:   frame[0],
		Type:      frame[1],
		Flags:     frame[2],
		Reserved:  frame[3],
		Length:    binary.BigEndian.Uint32(frame[4:8]),
		Checksum:  binary.BigEndian.Uint32(frame[8:12]),
		Timestamp: binary.BigEndian.Uint32(frame[12:16]),
	}

	// Validate frame version
	if header.Version != BinaryFrameVersion {
		return nil, fmt.Errorf("unsupported frame version: %d", header.Version)
	}

	// Validate frame length
	if len(frame) < BinaryFrameHeaderSize+int(header.Length) {
		return nil, fmt.Errorf("incomplete frame: got %d bytes, expected %d", len(frame), BinaryFrameHeaderSize+int(header.Length))
	}

	// Extract payload
	payload := frame[BinaryFrameHeaderSize : BinaryFrameHeaderSize+int(header.Length)]

	// Verify checksum
	if calculateCRC32(payload) != header.Checksum {
		return nil, fmt.Errorf("checksum mismatch")
	}

	// Decode payload based on frame flags and type
	var msg ActivationStreamMessage
	msg.Type = unmapMessageType(header.Type)
	msg.Timestamp = time.Unix(int64(header.Timestamp), 0)

	// Handle binary activation chunks (true binary framing without JSON/base64)
	if header.Type == mapMessageType(MsgTypeActivationChunk) {
		// Always use binary decoding for activation chunks (no JSON fallback)
		chunk, err := decodeBinaryActivationChunk(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to decode binary activation chunk: %w", err)
		}
		msg.Chunk = chunk
	} else {
		// Decode JSON payload for control messages only
		if err := json.Unmarshal(payload, &msg); err != nil {
			return nil, fmt.Errorf("failed to decode message payload: %w", err)
		}
		// Restore message type from binary code (in case JSON didn't have it)
		msg.Type = unmapMessageType(header.Type)
	}

	return &msg, nil
}

// encodeBinaryActivationChunk encodes activation chunk as pure binary (no JSON/base64)
func encodeBinaryActivationChunk(chunk *ActivationChunk) ([]byte, error) {
	inferenceIDBytes := []byte(chunk.Header.InferenceID)
	partitionIDBytes := []byte(chunk.Header.PartitionID)

	if len(inferenceIDBytes) > 255 || len(partitionIDBytes) > 255 {
		return nil, fmt.Errorf("inference ID or partition ID too long for binary encoding")
	}

	// Calculate total size: fixed fields + variable length strings + data
	totalSize := 1 + len(inferenceIDBytes) + 1 + len(partitionIDBytes) + 4 + 4 + 4 + 8 + 4 + len(chunk.Data)
	buf := make([]byte, 0, totalSize)

	// Encode inference ID
	buf = append(buf, uint8(len(inferenceIDBytes)))
	buf = append(buf, inferenceIDBytes...)

	// Encode partition ID
	buf = append(buf, uint8(len(partitionIDBytes)))
	buf = append(buf, partitionIDBytes...)

	// Encode fixed fields
	chunkIndexBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(chunkIndexBuf, uint32(chunk.ChunkIndex))
	buf = append(buf, chunkIndexBuf...)

	chunkSizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(chunkSizeBuf, uint32(chunk.ChunkSize))
	buf = append(buf, chunkSizeBuf...)

	sequenceNumBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(sequenceNumBuf, uint32(chunk.Header.SequenceNum))
	buf = append(buf, sequenceNumBuf...)

	timestampBuf := make([]byte, 8)
	binary.BigEndian.PutUint64(timestampBuf, uint64(chunk.Header.Timestamp.UnixNano()))
	buf = append(buf, timestampBuf...)

	// Calculate and encode data checksum
	dataChecksum := calculateCRC32(chunk.Data)
	checksumBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(checksumBuf, dataChecksum)
	buf = append(buf, checksumBuf...)

	// Append raw data
	buf = append(buf, chunk.Data...)

	return buf, nil
}

// decodeBinaryActivationChunk decodes binary activation chunk back to ActivationChunk
func decodeBinaryActivationChunk(data []byte) (*ActivationChunk, error) {
	if len(data) < 2 {
		return nil, fmt.Errorf("binary activation chunk too short")
	}

	offset := 0

	// Decode inference ID
	inferenceIDLen := int(data[offset])
	offset++
	if offset+inferenceIDLen > len(data) {
		return nil, fmt.Errorf("invalid inference ID length")
	}
	inferenceID := string(data[offset : offset+inferenceIDLen])
	offset += inferenceIDLen

	// Decode partition ID
	if offset >= len(data) {
		return nil, fmt.Errorf("missing partition ID length")
	}
	partitionIDLen := int(data[offset])
	offset++
	if offset+partitionIDLen > len(data) {
		return nil, fmt.Errorf("invalid partition ID length")
	}
	partitionID := string(data[offset : offset+partitionIDLen])
	offset += partitionIDLen

	// Check remaining data length (4+4+4+8+4 = 24 bytes minimum)
	if offset+24 > len(data) {
		return nil, fmt.Errorf("insufficient data for fixed fields")
	}

	// Decode fixed fields
	chunkIndex := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	chunkSize := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	sequenceNum := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	timestamp := binary.BigEndian.Uint64(data[offset : offset+8])
	offset += 8

	dataChecksum := binary.BigEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Extract raw data
	if offset > len(data) {
		return nil, fmt.Errorf("insufficient data for payload")
	}
	chunkData := data[offset:]

	// Verify data checksum
	if calculateCRC32(chunkData) != dataChecksum {
		return nil, fmt.Errorf("data checksum mismatch")
	}

	// Verify chunk size
	if uint32(len(chunkData)) != chunkSize {
		return nil, fmt.Errorf("chunk size mismatch: expected %d, got %d", chunkSize, len(chunkData))
	}

	// Reconstruct ActivationChunk
	chunk := &ActivationChunk{
		Header: ActivationHeader{
			InferenceID: inferenceID,
			PartitionID: partitionID,
			SequenceNum: int32(sequenceNum),
			Timestamp:   time.Unix(0, int64(timestamp)),
		},
		ChunkIndex: int32(chunkIndex),
		ChunkSize:  int32(chunkSize),
		Data:       chunkData,
		Checksum:   fmt.Sprintf("%08x", dataChecksum),
	}

	return chunk, nil
}

// calculateCRC32 calculates a simple CRC32 checksum
func calculateCRC32(data []byte) uint32 {
	// Simple implementation - in production would use proper CRC32
	var crc uint32 = 0xFFFFFFFF
	for _, b := range data {
		crc ^= uint32(b)
		for i := 0; i < 8; i++ {
			if crc&1 != 0 {
				crc = (crc >> 1) ^ 0xEDB88320
			} else {
				crc >>= 1
			}
		}
	}
	return ^crc
}
