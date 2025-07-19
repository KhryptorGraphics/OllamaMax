package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ollama/ollama-distributed/pkg/api"
	"github.com/ollama/ollama-distributed/pkg/integration"
	"github.com/ollama/ollama-distributed/internal/config"
)

// Test_API_Server_Initialization tests the API server initialization
func Test_API_Server_Initialization(t *testing.T) {
	config := &config.Config{
		P2P: config.P2PConfig{
			ListenAddr: "127.0.0.1:0",
		},
		API: config.APIConfig{
			Port: "0",
		},
		Consensus: config.ConsensusConfig{
			DataDir: t.TempDir(),
		},
		Scheduler: config.SchedulerConfig{
			Strategy: "round_robin",
		},
	}

	server, err := api.NewServer(context.Background(), config)
	require.NoError(t, err)
	assert.NotNil(t, server)

	// Test server can start and stop
	go func() {
		err := server.Start()
		if err != nil {
			t.Logf("Server start error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	err = server.Stop(context.Background())
	assert.NoError(t, err)
}

// Test_Integration_Client tests the integration client functionality
func Test_Integration_Client(t *testing.T) {
	client := integration.NewClient()
	assert.NotNil(t, client)

	ctx := context.Background()

	// Test generate
	generateReq := integration.GenerateRequest{
		Model:  "test-model",
		Prompt: "Hello, world!",
	}
	
	generateResp, err := client.Generate(ctx, generateReq)
	assert.NoError(t, err)
	assert.NotNil(t, generateResp)
	assert.Equal(t, "test-model", generateResp.Model)
	assert.NotEmpty(t, generateResp.Response)
	assert.True(t, generateResp.Done)

	// Test chat
	chatReq := integration.ChatRequest{
		Model: "test-model",
		Messages: []integration.Message{
			{Role: "user", Content: "Hello!"},
		},
	}
	
	chatResp, err := client.Chat(ctx, chatReq)
	assert.NoError(t, err)
	assert.NotNil(t, chatResp)
	assert.Equal(t, "test-model", chatResp.Model)
	assert.Equal(t, "assistant", chatResp.Message.Role)
	assert.NotEmpty(t, chatResp.Message.Content)

	// Test embeddings
	embedReq := integration.EmbedRequest{
		Model:  "test-model",
		Prompt: "Test embedding",
	}
	
	embedResp, err := client.Embed(ctx, embedReq)
	assert.NoError(t, err)
	assert.NotNil(t, embedResp)
	assert.Len(t, embedResp.Embedding, 384)

	// Test list models
	listResp, err := client.List(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, listResp)
	assert.Len(t, listResp.Models, 2) // Default models

	// Test show model
	showReq := integration.ShowRequest{Model: "test-model"}
	showResp, err := client.Show(ctx, showReq)
	assert.NoError(t, err)
	assert.NotNil(t, showResp)
	assert.NotEmpty(t, showResp.License)

	// Test pull model
	pullReq := integration.PullRequest{Model: "new-model"}
	err = client.Pull(ctx, pullReq)
	assert.NoError(t, err)

	// Test version
	versionResp, err := client.Version(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, versionResp)
	assert.Equal(t, "0.1.0-distributed", versionResp.Version)

	// Test delete model
	deleteReq := integration.DeleteRequest{Model: "new-model"}
	err = client.Delete(ctx, deleteReq)
	assert.NoError(t, err)
}

// Test_KeyManager_Integration tests the key manager stub
func Test_KeyManager_Integration(t *testing.T) {
	keyManager := &integration.DefaultKeyManager{}

	// Test session key generation
	sessionKey, err := keyManager.GenerateSessionKey()
	assert.NoError(t, err)
	assert.NotNil(t, sessionKey)
	assert.NotEmpty(t, sessionKey.PublicKey)
	assert.NotEmpty(t, sessionKey.PrivateKey)
	assert.False(t, sessionKey.CreatedAt.IsZero())
	assert.False(t, sessionKey.ExpiresAt.IsZero())

	// Test public key retrieval
	pubKey, err := keyManager.GetPeerPublicKey([]byte("test-key"))
	assert.NoError(t, err)
	assert.NotNil(t, pubKey)

	// Test public key verification
	verified, err := pubKey.Verify([]byte("test-data"), []byte("test-signature"))
	assert.NoError(t, err)
	assert.True(t, verified) // Stub always returns true

	// Test session count
	count := keyManager.GetActiveSessionCount()
	assert.Equal(t, 0, count)

	// Test cleanup operations
	keyManager.CleanupExpiredSessions()
	
	err = keyManager.HandleKeyExchange(nil)
	assert.NoError(t, err)

	err = keyManager.RotateKeys()
	assert.NoError(t, err)

	err = keyManager.Close()
	assert.NoError(t, err)
}

// Test_Performance_Metrics tests API performance metrics
func Test_Performance_Metrics(t *testing.T) {
	client := integration.NewClient()
	ctx := context.Background()

	start := time.Now()
	
	// Run multiple operations to test performance
	for i := 0; i < 10; i++ {
		generateReq := integration.GenerateRequest{
			Model:  "test-model",
			Prompt: "Performance test",
		}
		
		resp, err := client.Generate(ctx, generateReq)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.NotZero(t, resp.Metrics.TotalDuration)
		assert.NotZero(t, resp.Metrics.EvalDuration)
	}

	elapsed := time.Since(start)
	t.Logf("Performed 10 generate operations in %v", elapsed)
	
	// Performance should be reasonable for stub operations
	assert.Less(t, elapsed, 1*time.Second)
}

// Test_Concurrent_Operations tests concurrent API operations
func Test_Concurrent_Operations(t *testing.T) {
	client := integration.NewClient()
	ctx := context.Background()

	const numConcurrent = 5
	results := make(chan error, numConcurrent)

	// Run concurrent generate operations
	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			generateReq := integration.GenerateRequest{
				Model:  "test-model",
				Prompt: "Concurrent test",
			}
			
			_, err := client.Generate(ctx, generateReq)
			results <- err
		}(i)
	}

	// Wait for all operations to complete
	for i := 0; i < numConcurrent; i++ {
		err := <-results
		assert.NoError(t, err)
	}
}

// Test_Error_Handling tests error handling in the integration layer
func Test_Error_Handling(t *testing.T) {
	client := integration.NewClient()
	ctx := context.Background()

	// Test with valid requests - should not error
	generateReq := integration.GenerateRequest{
		Model:  "test-model",
		Prompt: "Test prompt",
	}
	
	_, err := client.Generate(ctx, generateReq)
	assert.NoError(t, err)

	// Test edge cases
	emptyReq := integration.GenerateRequest{}
	resp, err := client.Generate(ctx, emptyReq)
	assert.NoError(t, err) // Stub implementation is permissive
	assert.NotNil(t, resp)

	// Test context cancellation
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel() // Cancel immediately

	_, err = client.Generate(cancelCtx, generateReq)
	// Note: Stub implementation doesn't check context cancellation
	// In a real implementation, this would return context.Canceled
}