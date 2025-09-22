# Dockerfile for Machine Learning Workers
FROM python:3.11-slim

# Install system dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    g++ \
    cmake \
    libgomp1 \
    wget \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Install Python ML dependencies
RUN pip install --no-cache-dir \
    numpy==1.24.3 \
    scikit-learn==1.3.0 \
    pandas==2.0.3 \
    tensorflow==2.13.0 \
    torch==2.0.1 \
    transformers==4.30.2 \
    xgboost==1.7.6 \
    redis==4.6.0 \
    requests==2.31.0

WORKDIR /app

# Copy ML components
COPY src/ml/ ./src/ml/
COPY models/ ./models/
COPY config/ ./config/

# Create necessary directories
RUN mkdir -p /app/cache /app/logs /app/data

# Create healthcheck script
RUN echo '#!/usr/bin/env python3\n\
import sys\n\
import requests\n\
try:\n\
    response = requests.get("http://localhost:8081/health", timeout=5)\n\
    sys.exit(0 if response.status_code == 200 else 1)\n\
except:\n\
    sys.exit(1)' > healthcheck.py

# Create startup script
RUN echo '#!/bin/bash\n\
echo "Starting ML Worker..."\n\
python -u src/ml/worker.py' > start.sh && chmod +x start.sh

# Non-root user for security
RUN useradd -m -u 1001 mlworker && \
    chown -R mlworker:mlworker /app

USER mlworker

EXPOSE 8081

CMD ["./start.sh"]