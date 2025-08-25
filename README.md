# Streaming Cache

A **gRPC-based streaming cache service** backed by **Redis Cloud**.  
Enables real-time streaming writes and reads of ticker data (price, volume, timestamp), designed for trading and financial applications.

> **Note:** This service is not directly user-facing. It is intended for backend services (such as trading platforms) that interact with it via gRPC.

## Table of Contents
- [Functional Requirements](#functional-requirements)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [gRPC API](#grpc-api)
- [Quick Start](#quick-start)
- [Connecting to Redis Cloud](#connecting-to-redis-cloud)
- [Running with Docker Compose](#running-with-docker-compose)
- [Testing with Postman](#testing-with-postman)
- [Development](#development)

## Requirements

- **Streaming Writes:** Allow streaming writes into the cache.
- **Streaming Reads:** Allow streaming reads from the cache.
- **Concurrency:** Support multiple clients reading from and writing to the cache concurrently.
- **Low Latency:** Designed for real-time price feeds in trading applications.
- **Backend Integration:** Intended for use by backend services, not direct end users.

## Architecture

```
┌─────────────┐     gRPC     ┌─────────────┐    Redis     ┌───────────┐
│   Client    │<------------>│   Server    │<------------>│   Redis   │
│  (gRPC)     │              │  (Go)       │    Client    │  (Cloud)  │
└─────────────┘              └─────────────┘              └───────────┘
```

## Project Structure

```
.
├── client/             # gRPC client implementation (Go)
│   ├── main.go
│   └── Dockerfile
├── proto/              # Protocol Buffer definitions
│   └── cache/
│       ├── cache.proto
│       ├── cache.pb.go
│       └── cache_grpc.pb.go
├── server/             # gRPC server implementation (Go)
│   ├── main.go
│   ├── Dockerfile
│   └── docker-compose.yml
├── docker-compose.yml  # Main compose file (client+server)
├── .env                # Redis Cloud credentials (not committed)
└── README.md
```

## gRPC API

### Service Definition
```protobuf
service Cache {
    rpc Get(Tkr) returns (TkrData) {}
    rpc GetStream(Tkr) returns (stream TkrData) {}
    rpc Set(TkrData) returns (Ack) {}
    rpc SetStream(stream TkrData) returns (stream Ack) {}
}

message Tkr {
    string tkr = 1;
}

message TkrData {
    string tkr = 1;
    int64 timestamp = 2;
    double price = 3;
    int64 volume = 4;
}

message Ack {
    bool success = 1;
}
```

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.25+
- Redis 6.0+ (local or Redis Cloud)
- protoc (Protocol Buffer compiler)
- Postman (for API testing)

### 1. Clone the Repository
```bash
git clone <repository-url>
cd streaming-cache
```

### 2. Generate gRPC Code (optional)
```bash
cd proto
protoc --go_out=. --go-grpc_out=. cache/cache.proto
```

## Connecting to Redis Cloud

1. Create a `.env` file in the project root:
```env
REDIS_ADDR=your-redis-cloud-host:port
REDIS_PASSWORD=your-redis-password
```

2. The server will automatically connect to Redis Cloud on startup using these environment variables.

## Running with Docker Compose

### Option 1: Server Only
```bash
cd server
docker-compose up --build
```
This will start:
- Redis server
- gRPC server (port 50051)

### Option 2: Both Client and Server
From the project root:
```bash
docker-compose up --build
```

## Testing with Postman

1. Open Postman (v10+)
2. Create a new gRPC request
3. Import the proto file: `proto/cache/cache.proto`
4. Set target server to: `localhost:50051`

### Example Request (Set)
```json
{
  "tkr": "GOOGL",
  "timestamp": 1691234567,
  "price": 102.3,
  "volume": 100000
}
```

## Development

### Environment Variables

| Variable         | Description                   | Default             |
|------------------|------------------------------|---------------------|
| SERVER_ADDRESS   | gRPC server address          | :50051             |
| REDIS_ADDR       | Redis server address         | localhost:6379     |
| REDIS_PASSWORD   | Redis password (if required) | (empty)            |

### Running Locally

1. Start the server:
   ```bash
   cd server
   go run main.go
   ```
2. In another terminal, start the client:
   ```bash
   cd client
   go run main.go
   ```