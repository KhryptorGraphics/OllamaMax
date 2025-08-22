package multimodal

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TextProcessorImpl implements text processing capabilities
type TextProcessorImpl struct {
	modelPath      string
	maxTokens      int
	temperature    float64
	supportedTasks []string
}

// NewTextProcessorImpl creates a new text processor
func NewTextProcessorImpl(config *ProcessorConfig) *TextProcessorImpl {
	return &TextProcessorImpl{
		modelPath:   config.ModelPath,
		maxTokens:   2048,
		temperature: 0.7,
		supportedTasks: []string{
			"text_generation",
			"question_answering",
			"summarization",
			"translation",
			"sentiment_analysis",
		},
	}
}

// Process processes text inputs
func (tp *TextProcessorImpl) Process(ctx context.Context, inputs []Input, params map[string]interface{}) ([]Output, error) {
	outputs := make([]Output, 0, len(inputs))

	for _, input := range inputs {
		if input.Type != ModalityText {
			return nil, fmt.Errorf("expected text input, got %s", input.Type)
		}

		// Extract text from input data
		text := string(input.Data)
		if text == "" {
			return nil, fmt.Errorf("empty text input")
		}

		// Determine task from parameters or metadata
		task := "text_generation"
		if taskParam, ok := params["task"].(string); ok {
			task = taskParam
		}

		// Process based on task
		result, err := tp.processText(ctx, text, task, params)
		if err != nil {
			return nil, fmt.Errorf("text processing failed: %w", err)
		}

		output := Output{
			Type:       ModalityText,
			Data:       []byte(result),
			Format:     "text/plain",
			Confidence: tp.calculateConfidence(text, result),
			Metadata: map[string]interface{}{
				"task":            task,
				"input_length":    len(text),
				"output_length":   len(result),
				"processing_time": time.Since(time.Now()),
			},
			Timestamp: time.Now(),
		}

		outputs = append(outputs, output)
	}

	return outputs, nil
}

// GetSupportedFormats returns supported text formats
func (tp *TextProcessorImpl) GetSupportedFormats() []string {
	return []string{
		"text/plain",
		"text/markdown",
		"application/json",
	}
}

// GetCapabilities returns text processing capabilities
func (tp *TextProcessorImpl) GetCapabilities() []string {
	return tp.supportedTasks
}

// processText processes text based on the specified task
func (tp *TextProcessorImpl) processText(ctx context.Context, text, task string, params map[string]interface{}) (string, error) {
	switch task {
	case "text_generation":
		return tp.generateText(text, params)
	case "question_answering":
		return tp.answerQuestion(text, params)
	case "summarization":
		return tp.summarizeText(text, params)
	case "translation":
		return tp.translateText(text, params)
	case "sentiment_analysis":
		return tp.analyzeSentiment(text, params)
	default:
		return "", fmt.Errorf("unsupported task: %s", task)
	}
}

// generateText generates text continuation
func (tp *TextProcessorImpl) generateText(prompt string, params map[string]interface{}) (string, error) {
	// Simple text generation simulation
	// In a real implementation, this would call an actual language model

	maxTokens := tp.maxTokens
	if mt, ok := params["max_tokens"].(int); ok {
		maxTokens = mt
	}

	// Simulate text generation based on prompt
	words := strings.Fields(prompt)
	if len(words) == 0 {
		return "Generated text based on empty prompt.", nil
	}

	// Simple continuation based on last words
	lastWord := words[len(words)-1]

	var continuation string
	switch {
	case strings.Contains(strings.ToLower(prompt), "story"):
		continuation = fmt.Sprintf("Once upon a time, %s led to an amazing adventure...", lastWord)
	case strings.Contains(strings.ToLower(prompt), "code"):
		continuation = fmt.Sprintf("// %s implementation\nfunc %s() {\n    // TODO: implement\n}", lastWord, lastWord)
	case strings.Contains(strings.ToLower(prompt), "question"):
		continuation = fmt.Sprintf("The answer to your question about %s is...", lastWord)
	default:
		continuation = fmt.Sprintf("Continuing from %s, we can explore various aspects and implications...", lastWord)
	}

	// Limit output length
	if len(continuation) > maxTokens*4 { // Rough token estimation
		continuation = continuation[:maxTokens*4] + "..."
	}

	return continuation, nil
}

// answerQuestion answers questions based on context
func (tp *TextProcessorImpl) answerQuestion(text string, params map[string]interface{}) (string, error) {
	// Simple question answering simulation

	// Extract question and context
	parts := strings.Split(text, "?")
	if len(parts) < 2 {
		return "I need a question to answer. Please provide a question ending with '?'", nil
	}

	question := strings.TrimSpace(parts[0]) + "?"
	_ = strings.TrimSpace(strings.Join(parts[1:], "")) // context for future use

	// Simple keyword-based answering
	questionLower := strings.ToLower(question)

	switch {
	case strings.Contains(questionLower, "what"):
		return fmt.Sprintf("Based on the context, %s refers to a concept that...", extractKeyword(question)), nil
	case strings.Contains(questionLower, "how"):
		return "The process involves several steps: 1) Analysis, 2) Implementation, 3) Validation.", nil
	case strings.Contains(questionLower, "why"):
		return "This occurs because of the underlying principles and mechanisms involved.", nil
	case strings.Contains(questionLower, "when"):
		return "This typically happens under specific conditions or timeframes.", nil
	case strings.Contains(questionLower, "where"):
		return "This can be found or occurs in various locations depending on the context.", nil
	default:
		return fmt.Sprintf("Regarding your question '%s', the answer depends on the specific context provided.", question), nil
	}
}

// summarizeText creates a summary of the input text
func (tp *TextProcessorImpl) summarizeText(text string, params map[string]interface{}) (string, error) {
	// Simple summarization simulation

	sentences := strings.Split(text, ".")
	if len(sentences) <= 2 {
		return text, nil // Already short enough
	}

	// Extract key sentences (first, middle, last)
	var summary []string

	// First sentence
	if len(sentences) > 0 && strings.TrimSpace(sentences[0]) != "" {
		summary = append(summary, strings.TrimSpace(sentences[0]))
	}

	// Middle sentence
	if len(sentences) > 2 {
		mid := len(sentences) / 2
		if strings.TrimSpace(sentences[mid]) != "" {
			summary = append(summary, strings.TrimSpace(sentences[mid]))
		}
	}

	// Last meaningful sentence
	for i := len(sentences) - 1; i >= 0; i-- {
		if strings.TrimSpace(sentences[i]) != "" {
			summary = append(summary, strings.TrimSpace(sentences[i]))
			break
		}
	}

	return strings.Join(summary, ". ") + ".", nil
}

// translateText translates text to target language
func (tp *TextProcessorImpl) translateText(text string, params map[string]interface{}) (string, error) {
	// Simple translation simulation

	targetLang := "english"
	if tl, ok := params["target_language"].(string); ok {
		targetLang = tl
	}

	// Simple keyword replacement for demonstration
	translations := map[string]map[string]string{
		"spanish": {
			"hello": "hola",
			"world": "mundo",
			"good":  "bueno",
			"day":   "día",
		},
		"french": {
			"hello": "bonjour",
			"world": "monde",
			"good":  "bon",
			"day":   "jour",
		},
	}

	if targetLang == "english" {
		return text, nil // Already in English
	}

	if langMap, exists := translations[targetLang]; exists {
		result := text
		for english, translated := range langMap {
			result = strings.ReplaceAll(strings.ToLower(result), english, translated)
		}
		return result, nil
	}

	return fmt.Sprintf("[Translated to %s]: %s", targetLang, text), nil
}

// analyzeSentiment analyzes the sentiment of text
func (tp *TextProcessorImpl) analyzeSentiment(text string, params map[string]interface{}) (string, error) {
	// Simple sentiment analysis simulation

	textLower := strings.ToLower(text)

	positiveWords := []string{"good", "great", "excellent", "amazing", "wonderful", "fantastic", "love", "like", "happy", "joy"}
	negativeWords := []string{"bad", "terrible", "awful", "horrible", "hate", "dislike", "sad", "angry", "disappointed", "frustrated"}

	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(textLower, word)
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(textLower, word)
	}

	var sentiment string
	var confidence float64

	if positiveCount > negativeCount {
		sentiment = "positive"
		confidence = float64(positiveCount) / float64(positiveCount+negativeCount+1)
	} else if negativeCount > positiveCount {
		sentiment = "negative"
		confidence = float64(negativeCount) / float64(positiveCount+negativeCount+1)
	} else {
		sentiment = "neutral"
		confidence = 0.5
	}

	return fmt.Sprintf(`{
		"sentiment": "%s",
		"confidence": %.2f,
		"positive_indicators": %d,
		"negative_indicators": %d
	}`, sentiment, confidence, positiveCount, negativeCount), nil
}

// calculateConfidence calculates processing confidence
func (tp *TextProcessorImpl) calculateConfidence(input, output string) float64 {
	// Simple confidence calculation based on output length and input relevance
	if len(output) == 0 {
		return 0.0
	}

	if len(input) == 0 {
		return 0.5
	}

	// Base confidence on output/input ratio and some heuristics
	ratio := float64(len(output)) / float64(len(input))

	// Normalize ratio to 0-1 range
	confidence := 0.5 + (ratio-1.0)*0.1
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

// extractKeyword extracts a key word from a question
func extractKeyword(question string) string {
	words := strings.Fields(strings.ToLower(question))

	// Skip common question words
	skipWords := map[string]bool{
		"what": true, "how": true, "why": true, "when": true, "where": true,
		"is": true, "are": true, "was": true, "were": true,
		"the": true, "a": true, "an": true, "and": true, "or": true,
	}

	for _, word := range words {
		cleaned := strings.Trim(word, "?.,!;:")
		if len(cleaned) > 2 && !skipWords[cleaned] {
			return cleaned
		}
	}

	return "concept"
}
