package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LoggerConfig configures the structured logger
type LoggerConfig struct {
	// Basic configuration
	Level  LogLevel
	Format LogFormat
	Output io.Writer

	// File logging
	EnableFileLogging bool
	LogFilePath       string
	MaxFileSize       int64 // bytes
	MaxBackups        int
	MaxAge            int // days
	Compress          bool

	// Structured logging
	EnableStructured bool
	EnableCaller     bool
	EnableStackTrace bool

	// Performance
	BufferSize    int
	FlushInterval time.Duration

	// Context
	ServiceName    string
	ServiceVersion string
	Environment    string

	// Sampling
	EnableSampling bool
	SampleRate     float64
}

// LogFormat represents the log output format
type LogFormat string

const (
	FormatJSON    LogFormat = "json"
	FormatText    LogFormat = "text"
	FormatConsole LogFormat = "console"
)

// StructuredLogger provides structured logging capabilities
type StructuredLogger struct {
	config *LoggerConfig
	logger *slog.Logger

	// File rotation
	fileWriter *RotatingFileWriter

	// Buffering
	buffer *LogBuffer

	// Metrics
	metrics *LogMetrics

	// Context
	baseAttrs []slog.Attr

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time `json:"timestamp"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Service     string    `json:"service"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`

	// Context
	TraceID   string `json:"trace_id,omitempty"`
	SpanID    string `json:"span_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`

	// Source
	Caller   string `json:"caller,omitempty"`
	Function string `json:"function,omitempty"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`

	// Additional fields
	Fields map[string]interface{} `json:"fields,omitempty"`

	// Error details
	Error      string `json:"error,omitempty"`
	ErrorType  string `json:"error_type,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// LogMetrics tracks logging performance and statistics
type LogMetrics struct {
	TotalLogs   int64            `json:"total_logs"`
	LogsByLevel map[string]int64 `json:"logs_by_level"`
	ErrorCount  int64            `json:"error_count"`
	DroppedLogs int64            `json:"dropped_logs"`

	// Performance
	AverageLatency time.Duration `json:"average_latency"`
	BufferUsage    float64       `json:"buffer_usage"`
	FlushCount     int64         `json:"flush_count"`

	// File metrics
	FileSize     int64 `json:"file_size"`
	FilesRotated int64 `json:"files_rotated"`

	// Last updated
	LastUpdated time.Time `json:"last_updated"`

	mu sync.RWMutex
}

// NewStructuredLogger creates a new structured logger
func NewStructuredLogger(config *LoggerConfig) (*StructuredLogger, error) {
	if config == nil {
		config = &LoggerConfig{
			Level:            LevelInfo,
			Format:           FormatJSON,
			Output:           os.Stdout,
			EnableStructured: true,
			EnableCaller:     true,
			BufferSize:       1000,
			FlushInterval:    5 * time.Second,
			ServiceName:      "ollama-distributed",
			Environment:      "development",
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	sl := &StructuredLogger{
		config: config,
		metrics: &LogMetrics{
			LogsByLevel: make(map[string]int64),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	// Initialize base attributes
	sl.baseAttrs = []slog.Attr{
		slog.String("service", config.ServiceName),
		slog.String("version", config.ServiceVersion),
		slog.String("environment", config.Environment),
	}

	// Setup output writer
	var writer io.Writer = config.Output

	// Setup file logging if enabled
	if config.EnableFileLogging {
		fileWriter, err := NewRotatingFileWriter(&RotatingFileConfig{
			Filename:   config.LogFilePath,
			MaxSize:    config.MaxFileSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create file writer: %w", err)
		}
		sl.fileWriter = fileWriter

		// Use multi-writer for both console and file
		writer = io.MultiWriter(config.Output, fileWriter)
	}

	// Setup buffering if enabled
	if config.BufferSize > 0 {
		buffer := NewLogBuffer(config.BufferSize, config.FlushInterval, writer)
		sl.buffer = buffer
		writer = buffer
	}

	// Create slog logger
	var handler slog.Handler

	switch config.Format {
	case FormatJSON:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     slog.Level(config.Level),
			AddSource: config.EnableCaller,
		})
	case FormatText:
		handler = slog.NewTextHandler(writer, &slog.HandlerOptions{
			Level:     slog.Level(config.Level),
			AddSource: config.EnableCaller,
		})
	default:
		handler = slog.NewJSONHandler(writer, &slog.HandlerOptions{
			Level:     slog.Level(config.Level),
			AddSource: config.EnableCaller,
		})
	}

	// Add base attributes to handler
	for _, attr := range sl.baseAttrs {
		handler = handler.WithAttrs([]slog.Attr{attr})
	}

	sl.logger = slog.New(handler)

	// Start background tasks
	if sl.buffer != nil {
		sl.wg.Add(1)
		go sl.flushLoop()
	}

	sl.wg.Add(1)
	go sl.metricsLoop()

	return sl, nil
}

// Debug logs a debug message
func (sl *StructuredLogger) Debug(msg string, fields ...slog.Attr) {
	sl.log(LevelDebug, msg, fields...)
}

// Info logs an info message
func (sl *StructuredLogger) Info(msg string, fields ...slog.Attr) {
	sl.log(LevelInfo, msg, fields...)
}

// Warn logs a warning message
func (sl *StructuredLogger) Warn(msg string, fields ...slog.Attr) {
	sl.log(LevelWarn, msg, fields...)
}

// Error logs an error message
func (sl *StructuredLogger) Error(msg string, err error, fields ...slog.Attr) {
	attrs := fields
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		attrs = append(attrs, slog.String("error_type", fmt.Sprintf("%T", err)))

		if sl.config.EnableStackTrace {
			attrs = append(attrs, slog.String("stack_trace", getStackTrace()))
		}
	}
	sl.log(LevelError, msg, attrs...)
}

// Fatal logs a fatal message and returns a fatal error
// Note: This no longer calls os.Exit() - callers should handle the error appropriately
func (sl *StructuredLogger) Fatal(msg string, err error, fields ...slog.Attr) error {
	attrs := fields
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		attrs = append(attrs, slog.String("error_type", fmt.Sprintf("%T", err)))
		attrs = append(attrs, slog.String("stack_trace", getStackTrace()))
	}
	sl.log(LevelFatal, msg, attrs...)

	// Flush all buffers
	sl.Flush()

	// Return error instead of calling os.Exit()
	if err != nil {
		return fmt.Errorf("fatal error: %s: %w", msg, err)
	}
	return fmt.Errorf("fatal error: %s", msg)
}

// FatalAndExit logs a fatal message and exits the program
// This should only be used in main functions where immediate exit is required
func (sl *StructuredLogger) FatalAndExit(msg string, err error, fields ...slog.Attr) {
	attrs := fields
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
		attrs = append(attrs, slog.String("error_type", fmt.Sprintf("%T", err)))
		attrs = append(attrs, slog.String("stack_trace", getStackTrace()))
	}
	sl.log(LevelFatal, msg, attrs...)

	// Flush all buffers before exiting
	sl.Flush()
	os.Exit(1)
}

// WithContext returns a logger with context fields
func (sl *StructuredLogger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: sl,
		ctx:    ctx,
	}
}

// WithFields returns a logger with additional fields
func (sl *StructuredLogger) WithFields(fields ...slog.Attr) *FieldLogger {
	return &FieldLogger{
		logger: sl,
		fields: fields,
	}
}

// log performs the actual logging
func (sl *StructuredLogger) log(level LogLevel, msg string, fields ...slog.Attr) {
	start := time.Now()

	// Check sampling
	if sl.config.EnableSampling && !sl.shouldSample() {
		sl.updateMetrics(level, true, time.Since(start))
		return
	}

	// Add caller information if enabled
	if sl.config.EnableCaller {
		if pc, file, line, ok := runtime.Caller(2); ok {
			fields = append(fields, slog.String("caller", fmt.Sprintf("%s:%d", filepath.Base(file), line)))
			if fn := runtime.FuncForPC(pc); fn != nil {
				fields = append(fields, slog.String("function", fn.Name()))
			}
		}
	}

	// Convert slog.Attr to any for slog
	args := make([]any, len(fields))
	for i, field := range fields {
		args[i] = field
	}

	// Log using slog
	switch level {
	case LevelDebug:
		sl.logger.Debug(msg, args...)
	case LevelInfo:
		sl.logger.Info(msg, args...)
	case LevelWarn:
		sl.logger.Warn(msg, args...)
	case LevelError:
		sl.logger.Error(msg, args...)
	case LevelFatal:
		sl.logger.Error(msg, args...)
	}

	sl.updateMetrics(level, false, time.Since(start))
}

// shouldSample determines if a log should be sampled
func (sl *StructuredLogger) shouldSample() bool {
	// Simple sampling implementation
	// In production, you might want more sophisticated sampling
	return true // For now, log everything
}

// updateMetrics updates logging metrics
func (sl *StructuredLogger) updateMetrics(level LogLevel, dropped bool, latency time.Duration) {
	sl.metrics.mu.Lock()
	defer sl.metrics.mu.Unlock()

	if dropped {
		sl.metrics.DroppedLogs++
	} else {
		sl.metrics.TotalLogs++
		sl.metrics.LogsByLevel[level.String()]++

		if level == LevelError || level == LevelFatal {
			sl.metrics.ErrorCount++
		}

		// Update average latency
		if sl.metrics.TotalLogs == 1 {
			sl.metrics.AverageLatency = latency
		} else {
			sl.metrics.AverageLatency = (sl.metrics.AverageLatency + latency) / 2
		}
	}

	sl.metrics.LastUpdated = time.Now()
}

// flushLoop periodically flushes the log buffer
func (sl *StructuredLogger) flushLoop() {
	defer sl.wg.Done()

	if sl.buffer == nil {
		return
	}

	ticker := time.NewTicker(sl.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-sl.ctx.Done():
			sl.buffer.Flush()
			return
		case <-ticker.C:
			sl.buffer.Flush()
			sl.metrics.mu.Lock()
			sl.metrics.FlushCount++
			sl.metrics.mu.Unlock()
		}
	}
}

// metricsLoop updates metrics periodically
func (sl *StructuredLogger) metricsLoop() {
	defer sl.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sl.ctx.Done():
			return
		case <-ticker.C:
			sl.updateFileMetrics()
		}
	}
}

// updateFileMetrics updates file-related metrics
func (sl *StructuredLogger) updateFileMetrics() {
	if sl.fileWriter == nil {
		return
	}

	sl.metrics.mu.Lock()
	defer sl.metrics.mu.Unlock()

	// Update file size and rotation metrics
	sl.metrics.FileSize = sl.fileWriter.GetSize()
	sl.metrics.FilesRotated = sl.fileWriter.GetRotationCount()

	// Update buffer usage if buffering is enabled
	if sl.buffer != nil {
		sl.metrics.BufferUsage = sl.buffer.GetUsage()
	}
}

// Flush flushes all pending log entries
func (sl *StructuredLogger) Flush() {
	if sl.buffer != nil {
		sl.buffer.Flush()
	}
	if sl.fileWriter != nil {
		sl.fileWriter.Flush()
	}
}

// GetMetrics returns current logging metrics
func (sl *StructuredLogger) GetMetrics() *LogMetrics {
	sl.metrics.mu.RLock()
	defer sl.metrics.mu.RUnlock()

	// Create a copy of metrics
	metrics := &LogMetrics{
		TotalLogs:      sl.metrics.TotalLogs,
		LogsByLevel:    make(map[string]int64),
		ErrorCount:     sl.metrics.ErrorCount,
		DroppedLogs:    sl.metrics.DroppedLogs,
		AverageLatency: sl.metrics.AverageLatency,
		BufferUsage:    sl.metrics.BufferUsage,
		FlushCount:     sl.metrics.FlushCount,
		FileSize:       sl.metrics.FileSize,
		FilesRotated:   sl.metrics.FilesRotated,
		LastUpdated:    sl.metrics.LastUpdated,
	}

	for level, count := range sl.metrics.LogsByLevel {
		metrics.LogsByLevel[level] = count
	}

	return metrics
}

// Close closes the logger and cleans up resources
func (sl *StructuredLogger) Close() error {
	sl.cancel()
	sl.wg.Wait()

	sl.Flush()

	if sl.fileWriter != nil {
		return sl.fileWriter.Close()
	}

	return nil
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// ContextLogger wraps the structured logger with context
type ContextLogger struct {
	logger *StructuredLogger
	ctx    context.Context
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(msg string, fields ...slog.Attr) {
	fields = cl.addContextFields(fields)
	cl.logger.Debug(msg, fields...)
}

// Info logs an info message with context
func (cl *ContextLogger) Info(msg string, fields ...slog.Attr) {
	fields = cl.addContextFields(fields)
	cl.logger.Info(msg, fields...)
}

// Warn logs a warning message with context
func (cl *ContextLogger) Warn(msg string, fields ...slog.Attr) {
	fields = cl.addContextFields(fields)
	cl.logger.Warn(msg, fields...)
}

// Error logs an error message with context
func (cl *ContextLogger) Error(msg string, err error, fields ...slog.Attr) {
	fields = cl.addContextFields(fields)
	cl.logger.Error(msg, err, fields...)
}

// addContextFields extracts fields from context
func (cl *ContextLogger) addContextFields(fields []slog.Attr) []slog.Attr {
	// Extract common context values
	if traceID := cl.ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			fields = append(fields, slog.String("trace_id", id))
		}
	}

	if requestID := cl.ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			fields = append(fields, slog.String("request_id", id))
		}
	}

	if userID := cl.ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok {
			fields = append(fields, slog.String("user_id", id))
		}
	}

	return fields
}

// FieldLogger wraps the structured logger with additional fields
type FieldLogger struct {
	logger *StructuredLogger
	fields []slog.Attr
}

// Debug logs a debug message with additional fields
func (fl *FieldLogger) Debug(msg string, fields ...slog.Attr) {
	allFields := append(fl.fields, fields...)
	fl.logger.Debug(msg, allFields...)
}

// Info logs an info message with additional fields
func (fl *FieldLogger) Info(msg string, fields ...slog.Attr) {
	allFields := append(fl.fields, fields...)
	fl.logger.Info(msg, allFields...)
}

// Warn logs a warning message with additional fields
func (fl *FieldLogger) Warn(msg string, fields ...slog.Attr) {
	allFields := append(fl.fields, fields...)
	fl.logger.Warn(msg, allFields...)
}

// Error logs an error message with additional fields
func (fl *FieldLogger) Error(msg string, err error, fields ...slog.Attr) {
	allFields := append(fl.fields, fields...)
	fl.logger.Error(msg, err, allFields...)
}

// RotatingFileWriter handles log file rotation
type RotatingFileWriter struct {
	config        *RotatingFileConfig
	file          *os.File
	size          int64
	rotationCount int64
	mu            sync.Mutex
}

// RotatingFileConfig configures file rotation
type RotatingFileConfig struct {
	Filename   string
	MaxSize    int64 // bytes
	MaxBackups int
	MaxAge     int // days
	Compress   bool
}

// NewRotatingFileWriter creates a new rotating file writer
func NewRotatingFileWriter(config *RotatingFileConfig) (*RotatingFileWriter, error) {
	if config.MaxSize <= 0 {
		config.MaxSize = 100 * 1024 * 1024 // 100MB default
	}

	if config.MaxBackups <= 0 {
		config.MaxBackups = 5
	}

	if config.MaxAge <= 0 {
		config.MaxAge = 30 // 30 days default
	}

	rfw := &RotatingFileWriter{
		config: config,
	}

	// Create initial file
	if err := rfw.openFile(); err != nil {
		return nil, err
	}

	return rfw, nil
}

// Write writes data to the file, rotating if necessary
func (rfw *RotatingFileWriter) Write(p []byte) (n int, err error) {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	// Check if rotation is needed
	if rfw.size+int64(len(p)) > rfw.config.MaxSize {
		if err := rfw.rotate(); err != nil {
			return 0, err
		}
	}

	n, err = rfw.file.Write(p)
	rfw.size += int64(n)
	return n, err
}

// Flush flushes the file
func (rfw *RotatingFileWriter) Flush() error {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	if rfw.file != nil {
		return rfw.file.Sync()
	}
	return nil
}

// Close closes the file
func (rfw *RotatingFileWriter) Close() error {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()

	if rfw.file != nil {
		return rfw.file.Close()
	}
	return nil
}

// GetSize returns the current file size
func (rfw *RotatingFileWriter) GetSize() int64 {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()
	return rfw.size
}

// GetRotationCount returns the number of rotations
func (rfw *RotatingFileWriter) GetRotationCount() int64 {
	rfw.mu.Lock()
	defer rfw.mu.Unlock()
	return rfw.rotationCount
}

// openFile opens the log file
func (rfw *RotatingFileWriter) openFile() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(rfw.config.Filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open file
	file, err := os.OpenFile(rfw.config.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	// Get current size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}

	rfw.file = file
	rfw.size = info.Size()

	return nil
}

// rotate rotates the log file
func (rfw *RotatingFileWriter) rotate() error {
	// Close current file
	if rfw.file != nil {
		rfw.file.Close()
	}

	// Move current file to backup
	backupName := fmt.Sprintf("%s.%d", rfw.config.Filename, time.Now().Unix())
	if err := os.Rename(rfw.config.Filename, backupName); err != nil {
		return err
	}

	// Compress if enabled
	if rfw.config.Compress {
		go rfw.compressFile(backupName)
	}

	// Clean up old backups
	go rfw.cleanupOldFiles()

	// Open new file
	if err := rfw.openFile(); err != nil {
		return err
	}

	rfw.rotationCount++
	return nil
}

// compressFile compresses a log file
func (rfw *RotatingFileWriter) compressFile(filename string) {
	// Implementation would compress the file using gzip
	// For now, this is a placeholder
}

// cleanupOldFiles removes old log files
func (rfw *RotatingFileWriter) cleanupOldFiles() {
	// Implementation would clean up old files based on MaxBackups and MaxAge
	// For now, this is a placeholder
}

// LogBuffer provides buffered logging
type LogBuffer struct {
	buffer        [][]byte
	maxSize       int
	flushInterval time.Duration
	writer        io.Writer
	mu            sync.Mutex
	lastFlush     time.Time
}

// NewLogBuffer creates a new log buffer
func NewLogBuffer(maxSize int, flushInterval time.Duration, writer io.Writer) *LogBuffer {
	return &LogBuffer{
		buffer:        make([][]byte, 0, maxSize),
		maxSize:       maxSize,
		flushInterval: flushInterval,
		writer:        writer,
		lastFlush:     time.Now(),
	}
}

// Write writes data to the buffer
func (lb *LogBuffer) Write(p []byte) (n int, err error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Make a copy of the data
	data := make([]byte, len(p))
	copy(data, p)

	lb.buffer = append(lb.buffer, data)

	// Flush if buffer is full or interval has passed
	if len(lb.buffer) >= lb.maxSize || time.Since(lb.lastFlush) > lb.flushInterval {
		lb.flush()
	}

	return len(p), nil
}

// Flush flushes the buffer
func (lb *LogBuffer) Flush() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.flush()
}

// flush internal flush method (must be called with lock held)
func (lb *LogBuffer) flush() {
	if len(lb.buffer) == 0 {
		return
	}

	// Write all buffered data
	for _, data := range lb.buffer {
		lb.writer.Write(data)
	}

	// Clear buffer
	lb.buffer = lb.buffer[:0]
	lb.lastFlush = time.Now()
}

// GetUsage returns buffer usage as a percentage
func (lb *LogBuffer) GetUsage() float64 {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	return float64(len(lb.buffer)) / float64(lb.maxSize)
}
