package unit

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModelSyncManager tests the model synchronization manager
func TestModelSyncManager(t *testing.T) {
	// Use helper functions to create test components
	syncManager := createTestSyncManager(t)
	require.NotNil(t, syncManager)

	t.Run("TestInitialization", func(t *testing.T) {
		assert.NotNil(t, syncManager, "Sync manager should be created")
	})

	t.Run("TestSyncManagerStart", func(t *testing.T) {
		// Test that sync manager can be started
		err := syncManager.Start()
		assert.NoError(t, err, "Sync manager should start successfully")

		// Clean up
		defer syncManager.Shutdown(context.Background())
	})
}
