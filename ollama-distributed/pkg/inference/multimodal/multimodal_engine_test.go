package multimodal

import (
	"context"
	"testing"
	"time"
)

func TestMultiModalEngine(t *testing.T) {
	// Create multimodal engine configuration
	config := &MultiModalConfig{
		TextConfig: &ProcessorConfig{
			Enabled:    true,
			ModelPath:  "/models/text",
			BatchSize:  4,
			Timeout:    30 * time.Second,
			MaxWorkers: 2,
		},
		ImageConfig: &ProcessorConfig{
			Enabled:    true,
			ModelPath:  "/models/image",
			BatchSize:  2,
			Timeout:    60 * time.Second,
			MaxWorkers: 1,
		},
		AudioConfig: &ProcessorConfig{
			Enabled:    true,
			ModelPath:  "/models/audio",
			BatchSize:  1,
			Timeout:    45 * time.Second,
			MaxWorkers: 1,
		},
		VideoConfig: &ProcessorConfig{
			Enabled:    false, // Keep disabled for now
			ModelPath:  "/models/video",
			BatchSize:  1,
			Timeout:    120 * time.Second,
			MaxWorkers: 1,
		},
		FusionConfig: &FusionConfig{
			DefaultMode:     FusionLate,
			LearningEnabled: false,
			WeightDecay:     0.01,
			UpdateInterval:  time.Hour,
		},
		RoutingConfig: &RoutingConfig{
			LoadBalancing: "round_robin",
			HealthCheck:   true,
			CheckInterval: 30 * time.Second,
			Timeout:       10 * time.Second,
		},
	}

	// Create multimodal engine
	engine, err := NewMultiModalEngine(config)
	if err != nil {
		t.Fatalf("Failed to create multimodal engine: %v", err)
	}
	defer engine.cancel()

	// Test text processing
	t.Run("TextProcessing", func(t *testing.T) {
		testTextProcessing(t, engine)
	})

	// Test image processing
	t.Run("ImageProcessing", func(t *testing.T) {
		testImageProcessing(t, engine)
	})

	// Test audio processing
	t.Run("AudioProcessing", func(t *testing.T) {
		testAudioProcessing(t, engine)
	})

	// Test multimodal fusion
	t.Run("MultiModalFusion", func(t *testing.T) {
		testMultiModalFusion(t, engine)
	})
}

func testTextProcessing(t *testing.T, engine *MultiModalEngine) {
	// Test text generation
	t.Run("TextGeneration", func(t *testing.T) {
		request := &MultiModalRequest{
			RequestID: "test-text-gen",
			Inputs: map[ModalityType][]Input{
				ModalityText: {
					{
						Type:     ModalityText,
						Data:     []byte("Write a story about"),
						Format:   "text/plain",
						Metadata: map[string]interface{}{"task": "text_generation"},
					},
				},
			},
			ModelID:    "text-model-1",
			Task:       "text_generation",
			Parameters: map[string]interface{}{"max_tokens": 100},
			FusionMode: FusionNone,
			Priority:   1,
			Timeout:    30 * time.Second,
		}

		ctx := context.Background()
		response, err := engine.ProcessRequest(ctx, request)
		if err != nil {
			t.Fatalf("Text generation failed: %v", err)
		}

		// Verify response
		if response.RequestID != request.RequestID {
			t.Errorf("Expected request ID %s, got %s", request.RequestID, response.RequestID)
		}

		if len(response.Outputs[ModalityText]) == 0 {
			t.Error("Expected text output")
		}

		output := response.Outputs[ModalityText][0]
		if len(output.Data) == 0 {
			t.Error("Expected non-empty output data")
		}

		if output.Confidence <= 0 || output.Confidence > 1 {
			t.Errorf("Expected confidence between 0 and 1, got %f", output.Confidence)
		}
	})

	// Test question answering
	t.Run("QuestionAnswering", func(t *testing.T) {
		request := &MultiModalRequest{
			RequestID: "test-qa",
			Inputs: map[ModalityType][]Input{
				ModalityText: {
					{
						Type:     ModalityText,
						Data:     []byte("What is artificial intelligence? AI is a field of computer science."),
						Format:   "text/plain",
						Metadata: map[string]interface{}{"task": "question_answering"},
					},
				},
			},
			ModelID:    "text-model-1",
			Task:       "question_answering",
			Parameters: map[string]interface{}{"task": "question_answering"},
			FusionMode: FusionNone,
		}

		ctx := context.Background()
		response, err := engine.ProcessRequest(ctx, request)
		if err != nil {
			t.Fatalf("Question answering failed: %v", err)
		}

		if len(response.Outputs[ModalityText]) == 0 {
			t.Error("Expected text output for question answering")
		}
	})
}

func testImageProcessing(t *testing.T, engine *MultiModalEngine) {
	// Create sample image data
	imageData := make([]byte, 1024) // 1KB sample image
	for i := range imageData {
		imageData[i] = byte(i % 256)
	}

	// Test image classification
	t.Run("ImageClassification", func(t *testing.T) {
		request := &MultiModalRequest{
			RequestID: "test-image-class",
			Inputs: map[ModalityType][]Input{
				ModalityImage: {
					{
						Type:     ModalityImage,
						Data:     imageData,
						Format:   "image/jpeg",
						Metadata: map[string]interface{}{"task": "image_classification"},
					},
				},
			},
			ModelID:    "image-model-1",
			Task:       "image_classification",
			Parameters: map[string]interface{}{"task": "image_classification"},
			FusionMode: FusionNone,
		}

		ctx := context.Background()
		response, err := engine.ProcessRequest(ctx, request)
		if err != nil {
			t.Fatalf("Image classification failed: %v", err)
		}

		if len(response.Outputs[ModalityImage]) == 0 {
			t.Error("Expected output for image classification")
		}

		output := response.Outputs[ModalityImage][0]
		if output.Format != "application/json" {
			t.Errorf("Expected JSON format, got %s", output.Format)
		}
	})

	// Test image captioning
	t.Run("ImageCaptioning", func(t *testing.T) {
		request := &MultiModalRequest{
			RequestID: "test-image-caption",
			Inputs: map[ModalityType][]Input{
				ModalityImage: {
					{
						Type:     ModalityImage,
						Data:     imageData,
						Format:   "image/png",
						Metadata: map[string]interface{}{"task": "image_captioning"},
					},
				},
			},
			ModelID:    "image-model-1",
			Task:       "image_captioning",
			Parameters: map[string]interface{}{"task": "image_captioning"},
			FusionMode: FusionNone,
		}

		ctx := context.Background()
		response, err := engine.ProcessRequest(ctx, request)
		if err != nil {
			t.Fatalf("Image captioning failed: %v", err)
		}

		if len(response.Outputs[ModalityImage]) == 0 {
			t.Error("Expected output for image captioning")
		}
	})
}

func testAudioProcessing(t *testing.T, engine *MultiModalEngine) {
	// Create sample audio data
	audioData := make([]byte, 16000) // 1 second of 16kHz audio
	for i := range audioData {
		audioData[i] = byte(i % 256)
	}

	// Test speech recognition
	t.Run("SpeechRecognition", func(t *testing.T) {
		request := &MultiModalRequest{
			RequestID: "test-speech-rec",
			Inputs: map[ModalityType][]Input{
				ModalityAudio: {
					{
						Type:     ModalityAudio,
						Data:     audioData,
						Format:   "audio/wav",
						Metadata: map[string]interface{}{"task": "speech_recognition"},
					},
				},
			},
			ModelID:    "audio-model-1",
			Task:       "speech_recognition",
			Parameters: map[string]interface{}{"task": "speech_recognition"},
			FusionMode: FusionNone,
		}

		ctx := context.Background()
		response, err := engine.ProcessRequest(ctx, request)
		if err != nil {
			t.Fatalf("Speech recognition failed: %v", err)
		}

		if len(response.Outputs[ModalityAudio]) == 0 {
			t.Error("Expected output for speech recognition")
		}

		output := response.Outputs[ModalityAudio][0]
		if output.Format != "application/json" {
			t.Errorf("Expected JSON format, got %s", output.Format)
		}
	})

	// Test audio classification
	t.Run("AudioClassification", func(t *testing.T) {
		request := &MultiModalRequest{
			RequestID: "test-audio-class",
			Inputs: map[ModalityType][]Input{
				ModalityAudio: {
					{
						Type:     ModalityAudio,
						Data:     audioData,
						Format:   "audio/mp3",
						Metadata: map[string]interface{}{"task": "audio_classification"},
					},
				},
			},
			ModelID:    "audio-model-1",
			Task:       "audio_classification",
			Parameters: map[string]interface{}{"task": "audio_classification"},
			FusionMode: FusionNone,
		}

		ctx := context.Background()
		response, err := engine.ProcessRequest(ctx, request)
		if err != nil {
			t.Fatalf("Audio classification failed: %v", err)
		}

		if len(response.Outputs[ModalityAudio]) == 0 {
			t.Error("Expected output for audio classification")
		}
	})
}

func testMultiModalFusion(t *testing.T, engine *MultiModalEngine) {
	// Create sample data for multiple modalities
	textData := []byte("Describe this image and audio")
	imageData := make([]byte, 2048)
	audioData := make([]byte, 8000)

	for i := range imageData {
		imageData[i] = byte(i % 256)
	}
	for i := range audioData {
		audioData[i] = byte(i % 128)
	}

	// Test multimodal fusion
	request := &MultiModalRequest{
		RequestID: "test-multimodal-fusion",
		Inputs: map[ModalityType][]Input{
			ModalityText: {
				{
					Type:   ModalityText,
					Data:   textData,
					Format: "text/plain",
				},
			},
			ModalityImage: {
				{
					Type:   ModalityImage,
					Data:   imageData,
					Format: "image/jpeg",
				},
			},
			ModalityAudio: {
				{
					Type:   ModalityAudio,
					Data:   audioData,
					Format: "audio/wav",
				},
			},
		},
		ModelID:    "multimodal-model-1",
		Task:       "multimodal_analysis",
		Parameters: map[string]interface{}{},
		FusionMode: FusionLate, // Test late fusion
		Priority:   1,
		Timeout:    60 * time.Second,
	}

	ctx := context.Background()
	response, err := engine.ProcessRequest(ctx, request)
	if err != nil {
		t.Fatalf("Multimodal fusion failed: %v", err)
	}

	// Verify we have outputs for each modality
	if len(response.Outputs[ModalityText]) == 0 {
		t.Error("Expected text output")
	}
	if len(response.Outputs[ModalityImage]) == 0 {
		t.Error("Expected image output")
	}
	if len(response.Outputs[ModalityAudio]) == 0 {
		t.Error("Expected audio output")
	}

	// Verify processing time is reasonable
	if response.ProcessingTime > 10*time.Second {
		t.Errorf("Processing time too long: %v", response.ProcessingTime)
	}

	// Verify overall confidence
	if response.Confidence <= 0 || response.Confidence > 1 {
		t.Errorf("Expected confidence between 0 and 1, got %f", response.Confidence)
	}
}

func TestTextProcessor(t *testing.T) {
	config := &ProcessorConfig{
		Enabled:   true,
		ModelPath: "/test/model",
		BatchSize: 1,
		Timeout:   30 * time.Second,
	}

	processor := NewTextProcessorImpl(config)

	// Test supported formats
	formats := processor.GetSupportedFormats()
	if len(formats) == 0 {
		t.Error("Expected supported formats")
	}

	// Test capabilities
	capabilities := processor.GetCapabilities()
	if len(capabilities) == 0 {
		t.Error("Expected capabilities")
	}

	// Test text processing
	inputs := []Input{
		{
			Type:   ModalityText,
			Data:   []byte("Test input text"),
			Format: "text/plain",
		},
	}

	params := map[string]interface{}{
		"task": "text_generation",
	}

	ctx := context.Background()
	outputs, err := processor.Process(ctx, inputs, params)
	if err != nil {
		t.Fatalf("Text processing failed: %v", err)
	}

	if len(outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(outputs))
	}

	output := outputs[0]
	if output.Type != ModalityText {
		t.Errorf("Expected text output, got %s", output.Type)
	}

	if len(output.Data) == 0 {
		t.Error("Expected non-empty output data")
	}
}

func BenchmarkMultiModalEngine(b *testing.B) {
	config := &MultiModalConfig{
		TextConfig: &ProcessorConfig{
			Enabled:    true,
			BatchSize:  1,
			Timeout:    10 * time.Second,
			MaxWorkers: 1,
		},
		ImageConfig: &ProcessorConfig{
			Enabled:    true,
			BatchSize:  1,
			Timeout:    10 * time.Second,
			MaxWorkers: 1,
		},
		AudioConfig: &ProcessorConfig{
			Enabled:    true,
			BatchSize:  1,
			Timeout:    10 * time.Second,
			MaxWorkers: 1,
		},
		FusionConfig: &FusionConfig{
			DefaultMode: FusionNone,
		},
		RoutingConfig: &RoutingConfig{
			LoadBalancing: "round_robin",
		},
	}

	engine, err := NewMultiModalEngine(config)
	if err != nil {
		b.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.cancel()

	request := &MultiModalRequest{
		RequestID: "bench-request",
		Inputs: map[ModalityType][]Input{
			ModalityText: {
				{
					Type:   ModalityText,
					Data:   []byte("Benchmark test input"),
					Format: "text/plain",
				},
			},
		},
		ModelID:    "bench-model",
		Task:       "text_generation",
		FusionMode: FusionNone,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		_, err := engine.ProcessRequest(ctx, request)
		if err != nil {
			b.Fatalf("Processing failed: %v", err)
		}
	}
}
