
## Streaming Cache

# Functional Requirements

Use case - latest price (candle movement) in a trading chart
1) Should allow streaming writes into the cache
2) Should allow streaming reads from the cache
3) Should allow multiple clients to read from and write to the cache concurrently


Details:
- gRPC server reads from and writes to redis cache
- It does not serve the actual user, instead user communicates with backend service via websocket and backend service communicates with cache service via gRPC
- Here scope is to implement a simple cache service that can be used by backend service to read and write data to cache



flow: 
```
gRPC client                        gRPC server
Cache Client - - - - - - - - - - > Cache server - - - - - - - - - - > Redis

```