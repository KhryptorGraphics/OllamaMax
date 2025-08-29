# Batch Processing

Process multiple requests efficiently with OllamaMax.

## Batch Processing Example

```python
import asyncio
import aiohttp

async def process_batch(prompts):
    async with aiohttp.ClientSession() as session:
        tasks = []
        for prompt in prompts:
            task = session.post('http://localhost:8081/api/generate', 
                              json={'model': 'llama2', 'prompt': prompt})
            tasks.append(task)
        responses = await asyncio.gather(*tasks)
        return responses
```