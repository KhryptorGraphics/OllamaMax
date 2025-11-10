/**
 * Message Queue Service
 * Handles queuing and processing of inference requests
 */

class MessageQueue {
  constructor() {
    this.queue = [];
    this.processing = false;
  }

  add(message, callback) {
    this.queue.push({ message, callback, timestamp: Date.now() });
    this.process();
  }

  async process() {
    if (this.processing || this.queue.length === 0) return;
    
    this.processing = true;
    
    while (this.queue.length > 0) {
      const { message, callback } = this.queue.shift();
      
      try {
        await callback(message);
      } catch (error) {
        console.error('Queue processing error:', error);
      }
      
      // Small delay between processing
      await new Promise(resolve => setTimeout(resolve, 100));
    }
    
    this.processing = false;
  }

  getLength() {
    return this.queue.length;
  }

  clear() {
    this.queue = [];
    this.processing = false;
  }

  getOldestTimestamp() {
    if (this.queue.length === 0) return null;
    return this.queue[0].timestamp;
  }

  getStats() {
    return {
      length: this.queue.length,
      processing: this.processing,
      oldestTimestamp: this.getOldestTimestamp()
    };
  }
}

module.exports = MessageQueue;

