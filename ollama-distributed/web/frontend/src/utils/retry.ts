/**
 * Retry utilities for robust API calls with exponential backoff
 * Used for handling network failures and rate limiting
 */

export interface RetryOptions {
  maxRetries?: number
  baseDelay?: number
  maxDelay?: number
  backoffFactor?: number
  retryCondition?: (error: any) => boolean
  onRetry?: (error: any, attempt: number) => void
}

const DEFAULT_OPTIONS: Required<RetryOptions> = {
  maxRetries: 3,
  baseDelay: 1000,
  maxDelay: 30000,
  backoffFactor: 2,
  retryCondition: (error) => {
    // Retry on network errors, timeouts, and 5xx server errors
    if (!error.response) return true // Network error
    const status = error.response.status
    return status >= 500 || status === 408 || status === 429
  },
  onRetry: () => {},
}

/**
 * Retry a function with exponential backoff
 */
export async function retryWithBackoff<T>(
  fn: () => Promise<T>,
  options: RetryOptions = {}
): Promise<T> {
  const opts = { ...DEFAULT_OPTIONS, ...options }
  let lastError: any

  for (let attempt = 0; attempt <= opts.maxRetries; attempt++) {
    try {
      return await fn()
    } catch (error) {
      lastError = error

      // Don't retry if this is the last attempt
      if (attempt === opts.maxRetries) {
        break
      }

      // Check if we should retry this error
      if (!opts.retryCondition(error)) {
        break
      }

      // Calculate delay with exponential backoff and jitter
      const delay = Math.min(
        opts.baseDelay * Math.pow(opts.backoffFactor, attempt),
        opts.maxDelay
      )
      
      // Add jitter to prevent thundering herd
      const jitteredDelay = delay + Math.random() * delay * 0.1

      opts.onRetry(error, attempt + 1)

      await sleep(jitteredDelay)
    }
  }

  throw lastError
}

/**
 * Retry with linear backoff
 */
export async function retryWithLinearBackoff<T>(
  fn: () => Promise<T>,
  options: Omit<RetryOptions, 'backoffFactor'> & { increment?: number } = {}
): Promise<T> {
  const { increment = 1000, ...baseOptions } = options
  const opts = { ...DEFAULT_OPTIONS, ...baseOptions }
  let lastError: any

  for (let attempt = 0; attempt <= opts.maxRetries; attempt++) {
    try {
      return await fn()
    } catch (error) {
      lastError = error

      if (attempt === opts.maxRetries) {
        break
      }

      if (!opts.retryCondition(error)) {
        break
      }

      const delay = Math.min(opts.baseDelay + (increment * attempt), opts.maxDelay)
      const jitteredDelay = delay + Math.random() * delay * 0.1

      opts.onRetry(error, attempt + 1)

      await sleep(jitteredDelay)
    }
  }

  throw lastError
}

/**
 * Circuit breaker pattern for preventing cascading failures
 */
export class CircuitBreaker<T extends (...args: any[]) => Promise<any>> {
  private failures = 0
  private lastFailTime = 0
  private state: 'closed' | 'open' | 'half-open' = 'closed'

  constructor(
    private fn: T,
    private options: {
      failureThreshold: number
      timeout: number
      resetTimeout: number
    } = {
      failureThreshold: 5,
      timeout: 60000, // 1 minute
      resetTimeout: 30000, // 30 seconds
    }
  ) {}

  async execute(...args: Parameters<T>): Promise<ReturnType<T>> {
    if (this.state === 'open') {
      if (Date.now() - this.lastFailTime > this.options.resetTimeout) {
        this.state = 'half-open'
      } else {
        throw new Error('Circuit breaker is open')
      }
    }

    try {
      const result = await this.fn(...args)
      
      if (this.state === 'half-open') {
        this.reset()
      }
      
      return result
    } catch (error) {
      this.recordFailure()
      throw error
    }
  }

  private recordFailure(): void {
    this.failures++
    this.lastFailTime = Date.now()

    if (this.failures >= this.options.failureThreshold) {
      this.state = 'open'
    }
  }

  private reset(): void {
    this.failures = 0
    this.state = 'closed'
  }

  getState(): string {
    return this.state
  }

  getFailures(): number {
    return this.failures
  }
}

/**
 * Timeout wrapper for promises
 */
export function withTimeout<T>(
  promise: Promise<T>,
  timeoutMs: number,
  timeoutMessage = 'Operation timed out'
): Promise<T> {
  return Promise.race([
    promise,
    new Promise<never>((_, reject) =>
      setTimeout(() => reject(new Error(timeoutMessage)), timeoutMs)
    ),
  ])
}

/**
 * Utility function to sleep for a given duration
 */
function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * Bulk retry for multiple operations
 */
export async function retryBulk<T>(
  operations: (() => Promise<T>)[],
  options: RetryOptions = {}
): Promise<(T | Error)[]> {
  return Promise.all(
    operations.map(async (op) => {
      try {
        return await retryWithBackoff(op, options)
      } catch (error) {
        return error instanceof Error ? error : new Error(String(error))
      }
    })
  )
}

/**
 * Rate-limited retry for handling rate limits
 */
export async function retryWithRateLimit<T>(
  fn: () => Promise<T>,
  options: RetryOptions & { rateLimitDelay?: number } = {}
): Promise<T> {
  const { rateLimitDelay = 5000, ...retryOptions } = options
  
  return retryWithBackoff(fn, {
    ...retryOptions,
    retryCondition: (error) => {
      const status = error.response?.status
      if (status === 429) return true // Rate limited
      return retryOptions.retryCondition?.(error) ?? DEFAULT_OPTIONS.retryCondition(error)
    },
    baseDelay: rateLimitDelay,
  })
}