from pkg.client.python.vault_client import VaultClient, Context, Message, Metadata
from datetime import datetime
import uuid

def main():
    # Create a client instance
    client = VaultClient()

    # Create a topic
    topic = client.create_topic("python-example", "Example topic for Python client")
    print(f"Created topic: {topic.name}")

    # Create a context
    context = Context(
        id=str(uuid.uuid4()),
        topic=topic.name,
        messages=[
            Message(role="user", content="Hello, Vault!"),
            Message(role="assistant", content="Hi there! How can I help you today?")
        ],
        metadata=Metadata(
            agent_id="example-agent",
            tags=["example", "python"],
            properties={"language": "python", "version": "1.0"}
        ),
        created_at=datetime.now(),
        updated_at=datetime.now()
    )

    # Store the context
    stored_context = client.store_context(context)
    print(f"Stored context with ID: {stored_context.id}")

    # Retrieve the context
    retrieved_context = client.get_context(topic.name, stored_context.id)
    print(f"Retrieved context: {retrieved_context.id}")
    print("Messages:")
    for msg in retrieved_context.messages:
        print(f"  {msg.role}: {msg.content}")

    # List all contexts in the topic
    contexts = client.list_contexts(topic.name)
    print(f"\nFound {len(contexts)} contexts in topic {topic.name}")

    # List all topics
    topics = client.list_topics()
    print(f"\nFound {len(topics)} topics:")
    for t in topics:
        print(f"  - {t.name}: {t.description}")

if __name__ == "__main__":
    main() 