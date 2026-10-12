"""Test suite for fraud scoring endpoints (/api/v1/fraud/score and /fraud/analyze)."""
from unittest.mock import MagicMock, patch
from fastapi.testclient import TestClient


def test_fraud_score_happy_path(client: TestClient):
    """Test successful fraud scoring with mocked Groq response."""
    mock_completion = MagicMock()
    mock_completion.choices = [
        MagicMock(message=MagicMock(content='{"risk_score": 0.25, "reason": "Low risk transaction"}'))
    ]
    
    with patch("main._groq") as mock_groq:
        mock_groq.chat.completions.create.return_value = mock_completion
        
        response = client.post(
            "/api/v1/fraud/score",
            json={"user_id": "usr_123", "amount": 500000}
        )
        
        assert response.status_code == 200
        data = response.json()
        assert "score" in data
        assert data["score"] == 0.25
        assert data["reason"] == "Low risk transaction"


def test_fraud_score_invalid_input(client: TestClient):
    """Test fraud scoring with invalid input types (missing amount or user_id) expecting 422."""
    response = client.post(
        "/api/v1/fraud/score",
        json={"user_id": "usr_123"}  # missing 'amount'
    )
    assert response.status_code == 422


def test_fraud_score_groq_error_returns_503(client: TestClient):
    """Test that Groq API exceptions are caught and return HTTP 503."""
    with patch("main._groq") as mock_groq:
        mock_groq.chat.completions.create.side_effect = Exception("Groq API error")
        
        response = client.post(
            "/api/v1/fraud/score",
            json={"user_id": "usr_123", "amount": 1000000}
        )
        
        assert response.status_code == 503
        assert "scoring-error" in response.json()["detail"]
