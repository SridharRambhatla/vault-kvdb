# Go-KVDB: Distributed Key-Value Database with Context Management

A distributed key-value database designed to serve as a context manager for large language models (LLMs). It provides efficient storage, retrieval, and caching of LLM contexts while supporting horizontal scaling through sharding.

## Features

- **Distributed Architecture**: Horizontal scaling through sharding
- **Context Management**: Store and retrieve LLM contexts efficiently
- **Caching**: LRU cache for frequently accessed contexts
- **Compression**: Data compression to save storage space
- **REST API**: Easy integration with any application
- **Ollama Integration**: Built-in support for Ollama LLM

## Getting Started

### Prerequisites

- Go 1.21 or later
- Ollama installed and running locally (for LLM integration)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/go-kvdb.git
   cd go-kvdb
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Build the project:
   ```bash
   go build
   ```

### Running the Server

The server can be run in two modes:

#### Single Node Mode

For development or small deployments:
```bash
go-kvdb -db-location=database/kvdb.db -http-addr=127.0.0.1:8080 -cache-size=1073741824
```

#### Sharded Mode

For production deployments with multiple nodes:

1. Configure shards in `sharding.toml`:
   ```toml
   [[shards]]
   name = "sh-1"
   idx = 0
   address = "127.0.0.1:8080"

   [[shards]]
   name = "sh-2"
   idx = 1
   address = "127.0.0.1:8081"

   [[shards]]
   name = "sh-3"
   idx = 2
   address = "127.0.0.1:8082"
   ```

2. Start all nodes using the provided scripts:
   - Windows: `launch.bat`
   - Linux/macOS: `./launch.sh`

   Or start nodes individually:
   ```bash
   # Node 1
   go-kvdb -db-location=database/sh-1.db -http-addr=127.0.0.1:8080 -config-file=sharding.toml -shard=sh-1 -cache-size=1073741824

   # Node 2
   go-kvdb -db-location=database/sh-2.db -http-addr=127.0.0.1:8081 -config-file=sharding.toml -shard=sh-2 -cache-size=1073741824

   # Node 3
   go-kvdb -db-location=database/sh-3.db -http-addr=127.0.0.1:8082 -config-file=sharding.toml -shard=sh-3 -cache-size=1073741824
   ```

### Configuration

- `-db-location`: Path to the database file
- `-http-addr`: HTTP server address and port
- `-config-file`: Path to sharding configuration file (required for sharded mode)
- `-shard`: Name of the current shard (required for sharded mode)
- `-cache-size`: Cache size in bytes (default: 1GB)

You can also set the cache size using the `KVDB_CACHE_SIZE` environment variable:
```bash
# Windows
set KVDB_CACHE_SIZE=2147483648  # 2GB
.\launch.bat

# Linux/macOS
KVDB_CACHE_SIZE=2147483648 ./launch.sh  # 2GB
```

## API Documentation

### Topics

#### Create Topic
```http
POST /api/v1/topics
Content-Type: application/json

{
    "name": "my-topic"
}
```

#### Get Topic
```http
GET /api/v1/topics/{topic-name}
```

#### List Topics
```http
GET /api/v1/topics
```

### Contexts

#### Store Context
```http
POST /api/v1/topics/{topic-name}/contexts
Content-Type: application/json

{
    "metadata": {
        "id": "context-1",
        "created_at": "2024-03-20T10:00:00Z",
        "model": "llama2",
        "tags": ["chat", "support"]
    },
    "content": {
        "text": "Previous conversation context...",
        "embeddings": [...]
    }
}
```

#### Get Context
```http
GET /api/v1/topics/{topic-name}/contexts/{context-id}
```

#### Delete Context
```http
DELETE /api/v1/topics/{topic-name}/contexts/{context-id}
```

#### List Contexts
```http
GET /api/v1/topics/{topic-name}/contexts
```

## Sharding

The database uses consistent hashing to distribute data across shards. Each key (topic name or context ID) is hashed to determine which shard should handle it. The system automatically redirects requests to the correct shard.

### Benefits

1. **Scalability**: Distribute data across multiple nodes
2. **Performance**: Each shard handles a subset of the data
3. **Fault Tolerance**: If one shard fails, others continue to operate
4. **Geographic Distribution**: Shards can be placed in different locations

### Adding New Shards

1. Add the new shard configuration to `sharding.toml`
2. Create a new database file for the shard
3. Start the new shard node
4. The system will automatically redistribute data

## Future Enhancements

### High Priority
- [ ] Implement data rebalancing when adding/removing shards
- [ ] Add replication for fault tolerance
- [ ] Implement shard health monitoring
- [ ] Add metrics and monitoring

### Medium Priority
- [ ] Support for more LLM providers
- [ ] Advanced caching strategies
- [ ] Query optimization
- [ ] Backup and restore functionality

## Security and Privacy

- [ ] Add authentication and authorization
- [ ] Implement data encryption
- [ ] Add rate limiting
- [ ] Implement audit logging

## Technical Stack

- **Language**: Go
- **Database**: BoltDB
- **Caching**: Custom LRU cache
- **API**: REST
- **Configuration**: TOML
- **LLM Integration**: Ollama

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details. 