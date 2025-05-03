import requests
import json
from typing import List, Optional, Dict, Any
from dataclasses import dataclass
from datetime import datetime

@dataclass
class Message:
    role: str
    content: str

@dataclass
class Metadata:
    agent_id: str
    tags: List[str]
    properties: Dict[str, Any]

@dataclass
class Context:
    id: str
    topic: str
    messages: List[Message]
    metadata: Metadata
    created_at: datetime
    updated_at: datetime

@dataclass
class Topic:
    name: str
    description: str
    created_at: datetime
    updated_at: datetime

class VaultClient:
    def __init__(self, base_url: str = "http://127.0.0.1:8080"):
        self.base_url = base_url.rstrip("/")
        self.session = requests.Session()

    def _make_request(self, method: str, endpoint: str, data: Optional[dict] = None) -> dict:
        url = f"{self.base_url}{endpoint}"
        response = self.session.request(method, url, json=data)
        response.raise_for_status()
        return response.json()

    def create_topic(self, name: str, description: str = "") -> Topic:
        """Create a new topic."""
        data = {
            "name": name,
            "description": description,
            "created_at": datetime.now().isoformat(),
            "updated_at": datetime.now().isoformat()
        }
        response = self._make_request("POST", "/api/v1/topics", data)
        return Topic(**response["data"])

    def get_topic(self, name: str) -> Optional[Topic]:
        """Get a topic by name."""
        data = {"name": name}
        response = self._make_request("POST", "/api/v1/topics/get", data)
        return Topic(**response["data"]) if response["data"] else None

    def list_topics(self) -> List[Topic]:
        """List all topics."""
        response = self._make_request("GET", "/api/v1/topics/list")
        return [Topic(**topic) for topic in response["data"]]

    def store_context(self, context: Context) -> Context:
        """Store a context."""
        data = {
            "id": context.id,
            "topic": context.topic,
            "messages": [{"role": m.role, "content": m.content} for m in context.messages],
            "metadata": {
                "agent_id": context.metadata.agent_id,
                "tags": context.metadata.tags,
                "properties": context.metadata.properties
            },
            "created_at": context.created_at.isoformat(),
            "updated_at": context.updated_at.isoformat()
        }
        response = self._make_request("POST", "/api/v1/topics/contexts/store", data)
        return Context(**response["data"])

    def get_context(self, topic: str, context_id: str) -> Optional[Context]:
        """Get a context by ID."""
        data = {
            "topic": topic,
            "context_id": context_id
        }
        response = self._make_request("POST", "/api/v1/topics/contexts/get", data)
        if not response["data"]:
            return None
        
        data = response["data"]
        return Context(
            id=data["id"],
            topic=data["topic"],
            messages=[Message(**msg) for msg in data["messages"]],
            metadata=Metadata(**data["metadata"]),
            created_at=datetime.fromisoformat(data["created_at"]),
            updated_at=datetime.fromisoformat(data["updated_at"])
        )

    def delete_context(self, topic: str, context_id: str) -> bool:
        """Delete a context."""
        data = {
            "topic": topic,
            "context_id": context_id
        }
        response = self._make_request("POST", "/api/v1/topics/contexts/delete", data)
        return response["success"]

    def list_contexts(self, topic: str) -> List[Context]:
        """List all contexts for a topic."""
        data = {"topic": topic}
        response = self._make_request("POST", "/api/v1/topics/contexts/list", data)
        contexts = []
        for ctx_data in response["data"]:
            contexts.append(Context(
                id=ctx_data["id"],
                topic=ctx_data["topic"],
                messages=[Message(**msg) for msg in ctx_data["messages"]],
                metadata=Metadata(**ctx_data["metadata"]),
                created_at=datetime.fromisoformat(ctx_data["created_at"]),
                updated_at=datetime.fromisoformat(ctx_data["updated_at"])
            ))
        return contexts 