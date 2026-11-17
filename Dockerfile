FROM golang:latest

# Install Python 3 and basic utilities
RUN apt-get update && \
    apt-get install -y --no-install-recommends python3 ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Cache Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose the card-ui server port
EXPOSE 8765

# Start card-ui HTTP server
CMD ["python3", "server.py", "--host", "0.0.0.0", "--port", "8765", "--no-open"]
