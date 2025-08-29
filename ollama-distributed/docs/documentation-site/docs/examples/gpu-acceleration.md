# GPU Acceleration

Enable GPU acceleration for improved performance.

## GPU Configuration

```yaml
# config.yaml
inference:
  gpu:
    enabled: true
    devices: [0, 1, 2, 3]
    memory_fraction: 0.9
```

Configure GPU support using the [Configuration Guide](../tutorial-basics/configuration.md).