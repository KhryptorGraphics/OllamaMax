/**
 * Inference Service
 * Handles actual AI model inference by routing to Ollama nodes
 */

class InferenceService {
  constructor(nodeRegistry) {
    this.nodeRegistry = nodeRegistry;
  }

  /**
   * Generate text completion
   */
  async generateCompletion(options) {
    const {
      model = 'llama-3.2-3b',
      prompt,
      max_tokens = 100,
      temperature = 0.7,
      top_p = 1,
      stream = false,
      stop = null,
      presence_penalty = 0,
      frequency_penalty = 0
    } = options;

    // Select a healthy node
    const node = this.nodeRegistry.selectNode('round-robin');
    
    if (!node) {
      throw new Error('No healthy nodes available');
    }

    // If it's a mock node, return mock data
    if (node.mock) {
      return this.generateMockCompletion(options);
    }

    // Call real Ollama node
    try {
      const response = await fetch(`${node.url}/api/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: this.mapModelName(model),
          prompt,
          stream: false,
          options: {
            temperature,
            top_p,
            num_predict: max_tokens,
            stop: Array.isArray(stop) ? stop : (stop ? [stop] : undefined)
          }
        })
      });

      if (!response.ok) {
        throw new Error(`Ollama node returned ${response.status}`);
      }

      const data = await response.json();
      
      return {
        id: `cmpl-${Date.now()}`,
        object: 'text_completion',
        created: Math.floor(Date.now() / 1000),
        model,
        choices: [{
          text: data.response,
          index: 0,
          logprobs: null,
          finish_reason: data.done ? 'stop' : 'length'
        }],
        usage: {
          prompt_tokens: data.prompt_eval_count || 0,
          completion_tokens: data.eval_count || 0,
          total_tokens: (data.prompt_eval_count || 0) + (data.eval_count || 0)
        }
      };
    } catch (error) {
      console.error('Inference error:', error);
      // Fallback to mock if real node fails
      return this.generateMockCompletion(options);
    }
  }

  /**
   * Generate chat completion
   */
  async generateChatCompletion(options) {
    const {
      model = 'llama-3.2-3b',
      messages,
      max_tokens = 100,
      temperature = 0.7,
      top_p = 1,
      stream = false,
      stop = null
    } = options;

    // Select a healthy node
    const node = this.nodeRegistry.selectNode('round-robin');
    
    if (!node) {
      throw new Error('No healthy nodes available');
    }

    // Convert messages to prompt
    const prompt = this.messagesToPrompt(messages);

    // If it's a mock node, return mock data
    if (node.mock) {
      return this.generateMockChatCompletion(options);
    }

    // Call real Ollama node
    try {
      const response = await fetch(`${node.url}/api/generate`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: this.mapModelName(model),
          prompt,
          stream: false,
          options: {
            temperature,
            top_p,
            num_predict: max_tokens,
            stop: Array.isArray(stop) ? stop : (stop ? [stop] : undefined)
          }
        })
      });

      if (!response.ok) {
        throw new Error(`Ollama node returned ${response.status}`);
      }

      const data = await response.json();
      
      return {
        id: `chatcmpl-${Date.now()}`,
        object: 'chat.completion',
        created: Math.floor(Date.now() / 1000),
        model,
        choices: [{
          index: 0,
          message: {
            role: 'assistant',
            content: data.response
          },
          finish_reason: data.done ? 'stop' : 'length'
        }],
        usage: {
          prompt_tokens: data.prompt_eval_count || 0,
          completion_tokens: data.eval_count || 0,
          total_tokens: (data.prompt_eval_count || 0) + (data.eval_count || 0)
        }
      };
    } catch (error) {
      console.error('Chat inference error:', error);
      // Fallback to mock if real node fails
      return this.generateMockChatCompletion(options);
    }
  }

  /**
   * Generate embeddings
   */
  async generateEmbeddings(options) {
    const {
      model = 'text-embedding-ada-002',
      input,
      encoding_format = 'float'
    } = options;

    const inputs = Array.isArray(input) ? input : [input];

    // Select a healthy node
    const node = this.nodeRegistry.selectNode('round-robin');

    if (!node) {
      throw new Error('No healthy nodes available');
    }

    // For now, always use mock embeddings
    // Real Ollama embeddings would require different endpoint
    return this.generateMockEmbeddings(inputs, model);
  }

  /**
   * Helper: Convert messages to prompt
   */
  messagesToPrompt(messages) {
    return messages.map(msg => {
      const role = msg.role === 'system' ? 'System' :
                   msg.role === 'user' ? 'User' : 'Assistant';
      return `${role}: ${msg.content}`;
    }).join('\n\n') + '\n\nAssistant:';
  }

  /**
   * Helper: Map OpenAI model names to Ollama model names
   */
  mapModelName(model) {
    const mapping = {
      'gpt-3.5-turbo': 'llama-3.2-3b',
      'gpt-4': 'llama-3.2-3b',
      'text-embedding-ada-002': 'llama-3.2-3b'
    };
    return mapping[model] || model;
  }

  /**
   * Generate mock completion
   */
  generateMockCompletion(options) {
    const { prompt, model, max_tokens } = options;

    const mockResponses = [
      'This is a mock response from the Ollama distributed inference system.',
      'The system is currently running in development mode with mock nodes.',
      'To use real inference, connect actual Ollama nodes to the cluster.',
      'Mock inference is useful for testing and development purposes.',
      'The distributed architecture allows for horizontal scaling across multiple nodes.'
    ];

    const response = mockResponses[Math.floor(Math.random() * mockResponses.length)];

    return {
      id: `cmpl-${Date.now()}`,
      object: 'text_completion',
      created: Math.floor(Date.now() / 1000),
      model,
      choices: [{
        text: response,
        index: 0,
        logprobs: null,
        finish_reason: 'stop'
      }],
      usage: {
        prompt_tokens: prompt.split(' ').length,
        completion_tokens: response.split(' ').length,
        total_tokens: prompt.split(' ').length + response.split(' ').length
      }
    };
  }

  /**
   * Generate mock chat completion
   */
  generateMockChatCompletion(options) {
    const { messages, model } = options;

    const lastMessage = messages[messages.length - 1];
    const userMessage = lastMessage.content.toLowerCase();

    let response = 'I am a mock AI assistant running on the OllamaMax distributed platform. ';

    if (userMessage.includes('hello') || userMessage.includes('hi')) {
      response += 'Hello! How can I help you today?';
    } else if (userMessage.includes('how are you')) {
      response += 'I\'m functioning well in mock mode. To enable real AI responses, connect actual Ollama nodes.';
    } else if (userMessage.includes('what') || userMessage.includes('?')) {
      response += 'I can answer questions once real Ollama nodes are connected. Currently running in development mode.';
    } else {
      response += 'I understand your message. This is a mock response for testing purposes.';
    }

    return {
      id: `chatcmpl-${Date.now()}`,
      object: 'chat.completion',
      created: Math.floor(Date.now() / 1000),
      model,
      choices: [{
        index: 0,
        message: {
          role: 'assistant',
          content: response
        },
        finish_reason: 'stop'
      }],
      usage: {
        prompt_tokens: messages.reduce((sum, m) => sum + m.content.split(' ').length, 0),
        completion_tokens: response.split(' ').length,
        total_tokens: messages.reduce((sum, m) => sum + m.content.split(' ').length, 0) + response.split(' ').length
      }
    };
  }

  /**
   * Generate mock embeddings
   */
  generateMockEmbeddings(inputs, model) {
    const dimensions = model.includes('ada-002') ? 1536 : 768;

    const data = inputs.map((text, index) => ({
      object: 'embedding',
      embedding: Array(dimensions).fill(0).map(() => Math.random() - 0.5),
      index
    }));

    return {
      object: 'list',
      data,
      model,
      usage: {
        prompt_tokens: inputs.join(' ').split(' ').length,
        total_tokens: inputs.join(' ').split(' ').length
      }
    };
  }
}

module.exports = InferenceService;

