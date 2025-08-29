package database

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// MockDatabase implements a test database for unit testing
type MockDatabase struct {
	data      map[string]interface{}
	connected bool
	latency   time.Duration
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		data:      make(map[string]interface{}),
		connected: false,
		latency:   time.Millisecond * 10,
	}
}

func (m *MockDatabase) Connect() error {
	time.Sleep(m.latency) // Simulate connection latency
	m.connected = true
	return nil
}

func (m *MockDatabase) Disconnect() error {
	m.connected = false
	return nil
}

func (m *MockDatabase) Get(key string) (interface{}, error) {
	if !m.connected {
		return nil, fmt.Errorf("database not connected")
	}
	time.Sleep(m.latency)
	
	value, exists := m.data[key]
	if !exists {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	return value, nil
}

func (m *MockDatabase) Set(key string, value interface{}) error {
	if !m.connected {
		return fmt.Errorf("database not connected")
	}
	time.Sleep(m.latency)
	m.data[key] = value
	return nil
}

func (m *MockDatabase) Delete(key string) error {
	if !m.connected {
		return fmt.Errorf("database not connected")
	}
	time.Sleep(m.latency)
	delete(m.data, key)
	return nil
}

func (m *MockDatabase) IsConnected() bool {
	return m.connected
}

func TestDatabaseConnection(t *testing.T) {
	db := NewMockDatabase()
	
	// Test initial state
	if db.IsConnected() {
		t.Error("Database should not be connected initially")
	}
	
	// Test connection
	err := db.Connect()
	if err != nil {
		t.Errorf("Failed to connect to database: %v", err)
	}
	
	if !db.IsConnected() {
		t.Error("Database should be connected after Connect()")
	}
	
	// Test disconnection
	err = db.Disconnect()
	if err != nil {
		t.Errorf("Failed to disconnect from database: %v", err)
	}
	
	if db.IsConnected() {
		t.Error("Database should not be connected after Disconnect()")
	}
}

func TestDatabaseCRUD(t *testing.T) {
	db := NewMockDatabase()
	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Disconnect()
	
	// Test Create/Set
	testKey := "test_key"
	testValue := "test_value"
	
	err = db.Set(testKey, testValue)
	if err != nil {
		t.Errorf("Failed to set value: %v", err)
	}
	
	// Test Read/Get
	retrievedValue, err := db.Get(testKey)
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}
	
	if retrievedValue != testValue {
		t.Errorf("Expected %v, got %v", testValue, retrievedValue)
	}
	
	// Test Update
	newValue := "updated_value"
	err = db.Set(testKey, newValue)
	if err != nil {
		t.Errorf("Failed to update value: %v", err)
	}
	
	retrievedValue, err = db.Get(testKey)
	if err != nil {
		t.Errorf("Failed to get updated value: %v", err)
	}
	
	if retrievedValue != newValue {
		t.Errorf("Expected %v, got %v", newValue, retrievedValue)
	}
	
	// Test Delete
	err = db.Delete(testKey)
	if err != nil {
		t.Errorf("Failed to delete value: %v", err)
	}
	
	_, err = db.Get(testKey)
	if err == nil {
		t.Error("Expected error when getting deleted key")
	}
}

func TestDatabaseErrorHandling(t *testing.T) {
	db := NewMockDatabase()
	
	// Test operations on disconnected database
	err := db.Set("key", "value")
	if err == nil {
		t.Error("Expected error when setting value on disconnected database")
	}
	
	_, err = db.Get("key")
	if err == nil {
		t.Error("Expected error when getting value from disconnected database")
	}
	
	err = db.Delete("key")
	if err == nil {
		t.Error("Expected error when deleting value from disconnected database")
	}
	
	// Connect and test non-existent key
	err = db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()
	
	_, err = db.Get("non_existent_key")
	if err == nil {
		t.Error("Expected error when getting non-existent key")
	}
}

func TestDatabaseTransaction(t *testing.T) {
	// Test transaction-like operations
	db := NewMockDatabase()
	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()
	
	// Simulate a transaction with multiple operations
	keys := []string{"tx_key1", "tx_key2", "tx_key3"}
	values := []string{"tx_value1", "tx_value2", "tx_value3"}
	
	// Batch insert
	for i, key := range keys {
		err := db.Set(key, values[i])
		if err != nil {
			t.Errorf("Transaction failed at key %s: %v", key, err)
		}
	}
	
	// Verify all values
	for i, key := range keys {
		value, err := db.Get(key)
		if err != nil {
			t.Errorf("Failed to get transaction key %s: %v", key, err)
		}
		if value != values[i] {
			t.Errorf("Transaction value mismatch for key %s: expected %s, got %s", key, values[i], value)
		}
	}
	
	// Batch cleanup
	for _, key := range keys {
		err := db.Delete(key)
		if err != nil {
			t.Errorf("Failed to cleanup transaction key %s: %v", key, err)
		}
	}
}

func TestDatabasePerformance(t *testing.T) {
	db := NewMockDatabase()
	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()
	
	// Test performance with multiple operations
	numOperations := 100
	start := time.Now()
	
	for i := 0; i < numOperations; i++ {
		key := fmt.Sprintf("perf_key_%d", i)
		value := fmt.Sprintf("perf_value_%d", i)
		
		err := db.Set(key, value)
		if err != nil {
			t.Errorf("Performance test failed at operation %d: %v", i, err)
		}
	}
	
	duration := time.Since(start)
	avgLatency := duration / time.Duration(numOperations)
	
	t.Logf("Performance test: %d operations in %v (avg: %v per operation)", 
		numOperations, duration, avgLatency)
	
	// Verify expected latency range (mock has 10ms latency)
	expectedMin := time.Millisecond * 8  // Allow some variance
	expectedMax := time.Millisecond * 15
	
	if avgLatency < expectedMin {
		t.Errorf("Average latency %v is suspiciously low (expected >%v)", avgLatency, expectedMin)
	}
	
	if avgLatency > expectedMax {
		t.Errorf("Average latency %v is too high (expected <%v)", avgLatency, expectedMax)
	}
}

func TestDatabaseConcurrency(t *testing.T) {
	db := NewMockDatabase()
	err := db.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()
	
	// Test concurrent access
	numGoroutines := 10
	numOperations := 50
	done := make(chan bool, numGoroutines)
	
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			defer func() { done <- true }()
			
			for i := 0; i < numOperations; i++ {
				key := fmt.Sprintf("concurrent_%d_%d", goroutineID, i)
				value := fmt.Sprintf("value_%d_%d", goroutineID, i)
				
				// Set value
				err := db.Set(key, value)
				if err != nil {
					t.Errorf("Concurrent set failed for goroutine %d, operation %d: %v", 
						goroutineID, i, err)
					return
				}
				
				// Get value
				retrievedValue, err := db.Get(key)
				if err != nil {
					t.Errorf("Concurrent get failed for goroutine %d, operation %d: %v", 
						goroutineID, i, err)
					return
				}
				
				if retrievedValue != value {
					t.Errorf("Concurrent value mismatch for goroutine %d, operation %d: expected %s, got %s", 
						goroutineID, i, value, retrievedValue)
					return
				}
				
				// Delete value
				err = db.Delete(key)
				if err != nil {
					t.Errorf("Concurrent delete failed for goroutine %d, operation %d: %v", 
						goroutineID, i, err)
					return
				}
			}
		}(g)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestDatabaseConnectionPool(t *testing.T) {
	// Test simulated connection pool behavior
	poolSize := 5
	connections := make([]*MockDatabase, poolSize)
	
	// Initialize connection pool
	for i := 0; i < poolSize; i++ {
		connections[i] = NewMockDatabase()
		err := connections[i].Connect()
		if err != nil {
			t.Errorf("Failed to connect pool connection %d: %v", i, err)
		}
	}
	
	// Test all connections
	for i, conn := range connections {
		if !conn.IsConnected() {
			t.Errorf("Pool connection %d is not connected", i)
		}
		
		// Test basic operations on each connection
		key := fmt.Sprintf("pool_test_%d", i)
		value := fmt.Sprintf("pool_value_%d", i)
		
		err := conn.Set(key, value)
		if err != nil {
			t.Errorf("Pool connection %d set operation failed: %v", i, err)
		}
		
		retrievedValue, err := conn.Get(key)
		if err != nil {
			t.Errorf("Pool connection %d get operation failed: %v", i, err)
		}
		
		if retrievedValue != value {
			t.Errorf("Pool connection %d value mismatch: expected %s, got %s", i, value, retrievedValue)
		}
	}
	
	// Cleanup connections
	for i, conn := range connections {
		err := conn.Disconnect()
		if err != nil {
			t.Errorf("Failed to disconnect pool connection %d: %v", i, err)
		}
	}
}

// Benchmark tests
func BenchmarkDatabaseSet(b *testing.B) {
	db := NewMockDatabase()
	db.latency = 0 // Remove artificial latency for benchmarking
	
	err := db.Connect()
	if err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := fmt.Sprintf("bench_value_%d", i)
		err := db.Set(key, value)
		if err != nil {
			b.Errorf("Benchmark set failed: %v", err)
		}
	}
}

func BenchmarkDatabaseGet(b *testing.B) {
	db := NewMockDatabase()
	db.latency = 0 // Remove artificial latency for benchmarking
	
	err := db.Connect()
	if err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()
	
	// Pre-populate data
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("bench_key_%d", i)
		value := fmt.Sprintf("bench_value_%d", i)
		db.Set(key, value)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_key_%d", i%1000)
		_, err := db.Get(key)
		if err != nil {
			b.Errorf("Benchmark get failed: %v", err)
		}
	}
}

func BenchmarkDatabaseDelete(b *testing.B) {
	db := NewMockDatabase()
	db.latency = 0 // Remove artificial latency for benchmarking
	
	err := db.Connect()
	if err != nil {
		b.Fatalf("Failed to connect: %v", err)
	}
	defer db.Disconnect()
	
	// Pre-populate data
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_delete_key_%d", i)
		value := fmt.Sprintf("bench_delete_value_%d", i)
		db.Set(key, value)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("bench_delete_key_%d", i)
		err := db.Delete(key)
		if err != nil {
			b.Errorf("Benchmark delete failed: %v", err)
		}
	}
}