# Vault

Vault is a lightweight, in-memory context management system designed specifically for local AI agents. It provides a simple way to store, retrieve, and manage conversation contexts before model inference.

## Features

- In-memory storage with LRU cache for fast context retrieval
- Topic-based organization of contexts
- RESTful API endpoints
- Python client library
- Thread-safe operations
- Automatic cache eviction based on size

## Use Cases

- Store conversation history for local LLMs
- Manage context windows for different topics
- Cache frequently accessed contexts
- Share contexts between multiple agents
- Pre-inference context preparation

## Installation

### Server

```bash
# Clone the repository
git clone https://github.com/yourusername/vault.git
cd vault

# Build the server
go build -o vault.exe cmd/vault/main.go
```

### Python Client

```bash
pip install requests
```

## Usage

### Starting the Server

```bash
./vault.exe --http-addr=127.0.0.1:8080 --cache-size=1073741824
```

### Using the Python Client

```python
from vault_client import VaultClient, Context, Message, Metadata
from datetime import datetime
import uuid

# Create a client instance
client = VaultClient()

# Create a topic for your AI agent
topic = client.create_topic("my-agent", "Conversation history for my AI agent")

# Store a context before inference
context = Context(
    id=str(uuid.uuid4()),
    topic=topic.name,
    messages=[
        Message(role="user", content="What is the capital of France?"),
        Message(role="assistant", content="The capital of France is Paris.")
    ],
    metadata=Metadata(
        agent_id="my-agent",
        tags=["geography", "capitals"],
        properties={"model": "local-llm", "temperature": "0.7"}
    ),
    created_at=datetime.now(),
    updated_at=datetime.now()
)

# Store the context
stored_context = client.store_context(context)

# Retrieve the context before inference
retrieved_context = client.get_context(topic.name, stored_context.id)
```

## API Endpoints

### Topics

- `POST /api/v1/topics` - Create a new topic
- `POST /api/v1/topics/get` - Get a topic by name
- `GET /api/v1/topics/list` - List all topics

### Contexts

- `POST /api/v1/topics/contexts/store` - Store a context
- `POST /api/v1/topics/contexts/get` - Get a context by ID
- `POST /api/v1/topics/contexts/delete` - Delete a context
- `POST /api/v1/topics/contexts/list` - List all contexts in a topic

## Response Format

All API responses follow this format:

```json
{
    "success": true,
    "data": {
        // Response data
    },
    "error": null
}
```

## Error Handling

The Python client raises exceptions for HTTP errors and invalid responses. The server returns appropriate HTTP status codes and error messages in the response body.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.