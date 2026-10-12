"""Pytest configuration and shared fixtures for backend-ai tests."""
import os
import pytest
from fastapi.testclient import TestClient

# Ensure test environment uses fallback when GROQ_API_KEY is not explicitly set
os.environ.setdefault("GROQ_API_KEY", "test-mock-key")

from main import app


@pytest.fixture
def client() -> TestClient:
    """Provide a FastAPI TestClient instance for testing endpoints."""
    return TestClient(app)
