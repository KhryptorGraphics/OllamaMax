/**
 * @fileoverview Models management API client
 * @description Handles model operations, distribution, and synchronization
 */

import { BaseAPIClient } from './base';
import {
  ModelInfo,
  DownloadProgress,
  SyncStatus,
  ModelRequest,
  PullRequest,
  GenerateRequest,
  ChatRequest,
  APIResponse,
  RequestConfig,
} from '../../types/api';

export class ModelsAPI extends BaseAPIClient {
  /**
   * List all available models
   */
  async list(config?: RequestConfig): Promise<ModelInfo[]> {
    const response = await this.get<{ models: ModelInfo[] }>('/api/tags', config);
    return response.data!.models;
  }

  /**
   * Get specific model information
   */
  async get(name: string, config?: RequestConfig): Promise<ModelInfo> {
    const response = await this.post<ModelInfo>('/api/show', { name }, config);
    return response.data!;
  }

  /**
   * Download/pull a model
   */
  async download(name: string, options?: Partial<PullRequest>, config?: RequestConfig): Promise<DownloadProgress> {
    const response = await this.post<DownloadProgress>(
      '/api/pull',
      { name, ...options },
      config
    );
    return response.data!;
  }

  /**
   * Stream model download progress
   */
  async *downloadStream(
    name: string,
    options?: Partial<PullRequest>,
    config?: RequestConfig
  ): AsyncGenerator<DownloadProgress> {
    const stream = await this.stream('/api/pull', {
      ...config,
      method: 'POST',
      body: JSON.stringify({ name, stream: true, ...options }),
      headers: {
        'Content-Type': 'application/json',
        ...config?.headers,
      },
    });

    const reader = stream.getReader();
    const decoder = new TextDecoder();

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n').filter(line => line.trim());

        for (const line of lines) {
          try {
            const progress: DownloadProgress = JSON.parse(line);
            yield progress;
          } catch (error) {
            console.warn('Failed to parse download progress:', line);
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }

  /**
   * Delete a model
   */
  async delete(name: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.delete<{ message: string }>(`/api/delete`, {
      ...config,
      body: JSON.stringify({ name }),
      headers: {
        'Content-Type': 'application/json',
        ...config?.headers,
      },
    });
    return response.data!;
  }

  /**
   * Copy a model
   */
  async copy(source: string, destination: string, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/api/copy',
      { source, destination },
      config
    );
    return response.data!;
  }

  /**
   * Create a model from a Modelfile
   */
  async create(
    name: string,
    modelfile: string,
    options?: { path?: string; stream?: boolean },
    config?: RequestConfig
  ): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/api/create',
      { name, modelfile, ...options },
      config
    );
    return response.data!;
  }

  /**
   * Push a model to a registry
   */
  async push(name: string, options?: { insecure?: boolean; stream?: boolean }, config?: RequestConfig): Promise<{ message: string }> {
    const response = await this.post<{ message: string }>(
      '/api/push',
      { name, ...options },
      config
    );
    return response.data!;
  }

  /**
   * Get model synchronization status
   */
  async getSyncStatus(config?: RequestConfig): Promise<SyncStatus> {
    const response = await this.get<SyncStatus>('/api/distributed/sync-status', config);
    return response.data!;
  }

  /**
   * Get distributed models information
   */
  async getDistributedModels(config?: RequestConfig): Promise<ModelInfo[]> {
    const response = await this.get<{ models: ModelInfo[] }>('/api/distributed/models', config);
    return response.data!.models;
  }

  /**
   * Generate text completion
   */
  async generate(request: GenerateRequest, config?: RequestConfig): Promise<any> {
    const response = await this.post('/api/generate', request, config);
    return response.data;
  }

  /**
   * Stream text generation
   */
  async *generateStream(request: GenerateRequest, config?: RequestConfig): AsyncGenerator<any> {
    const stream = await this.stream('/api/generate', {
      ...config,
      method: 'POST',
      body: JSON.stringify({ ...request, stream: true }),
      headers: {
        'Content-Type': 'application/json',
        ...config?.headers,
      },
    });

    const reader = stream.getReader();
    const decoder = new TextDecoder();

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n').filter(line => line.trim());

        for (const line of lines) {
          try {
            const data = JSON.parse(line);
            yield data;
          } catch (error) {
            console.warn('Failed to parse generation response:', line);
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }

  /**
   * Chat completion
   */
  async chat(request: ChatRequest, config?: RequestConfig): Promise<any> {
    const response = await this.post('/api/chat', request, config);
    return response.data;
  }

  /**
   * Stream chat completion
   */
  async *chatStream(request: ChatRequest, config?: RequestConfig): AsyncGenerator<any> {
    const stream = await this.stream('/api/chat', {
      ...config,
      method: 'POST',
      body: JSON.stringify({ ...request, stream: true }),
      headers: {
        'Content-Type': 'application/json',
        ...config?.headers,
      },
    });

    const reader = stream.getReader();
    const decoder = new TextDecoder();

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        const chunk = decoder.decode(value);
        const lines = chunk.split('\n').filter(line => line.trim());

        for (const line of lines) {
          try {
            const data = JSON.parse(line);
            yield data;
          } catch (error) {
            console.warn('Failed to parse chat response:', line);
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  }

  /**
   * Generate embeddings
   */
  async embed(
    model: string,
    prompt: string,
    options?: Record<string, any>,
    config?: RequestConfig
  ): Promise<{ embedding: number[] }> {
    const response = await this.post<{ embedding: number[] }>(
      '/api/embeddings',
      { model, prompt, options },
      config
    );
    return response.data!;
  }

  /**
   * Get running models/processes
   */
  async getRunning(config?: RequestConfig): Promise<Array<any>> {
    const response = await this.get<{ models: Array<any> }>('/api/ps', config);
    return response.data!.models;
  }

  /**
   * Get model version information
   */
  async getVersion(config?: RequestConfig): Promise<{ version: string }> {
    const response = await this.get<{ version: string }>('/api/version', config);
    return response.data!;
  }

  // OpenAI compatibility endpoints

  /**
   * OpenAI-compatible chat completions
   */
  async openaiChat(request: any, config?: RequestConfig): Promise<any> {
    const response = await this.post('/v1/chat/completions', request, config);
    return response.data;
  }

  /**
   * OpenAI-compatible completions
   */
  async openaiCompletion(request: any, config?: RequestConfig): Promise<any> {
    const response = await this.post('/v1/completions', request, config);
    return response.data;
  }

  /**
   * OpenAI-compatible embeddings
   */
  async openaiEmbeddings(request: any, config?: RequestConfig): Promise<any> {
    const response = await this.post('/v1/embeddings', request, config);
    return response.data;
  }

  /**
   * OpenAI-compatible models list
   */
  async openaiModels(config?: RequestConfig): Promise<any> {
    const response = await this.get('/v1/models', config);
    return response.data;
  }

  /**
   * OpenAI-compatible model info
   */
  async openaiModel(model: string, config?: RequestConfig): Promise<any> {
    const response = await this.get(`/v1/models/${model}`, config);
    return response.data;
  }

  /**
   * Search models by name or description
   */
  async search(query: string, config?: RequestConfig): Promise<ModelInfo[]> {
    const models = await this.list(config);
    return models.filter(model => 
      model.name.toLowerCase().includes(query.toLowerCase()) ||
      model.family?.toLowerCase().includes(query.toLowerCase())
    );
  }

  /**
   * Get model usage statistics
   */
  async getModelStats(name: string, config?: RequestConfig): Promise<{
    download_count: number;
    inference_count: number;
    last_used: string;
    total_size: number;
  }> {
    const model = await this.get(name, config);
    return model.metrics;
  }

  /**
   * Monitor model synchronization
   */
  async *monitorSync(intervalMs: number = 5000): AsyncGenerator<SyncStatus> {
    while (true) {
      try {
        const status = await this.getSyncStatus();
        yield status;
        await new Promise(resolve => setTimeout(resolve, intervalMs));
      } catch (error) {
        console.error('Error monitoring sync status:', error);
        await new Promise(resolve => setTimeout(resolve, intervalMs * 2));
      }
    }
  }
}