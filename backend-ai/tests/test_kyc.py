"""Test suite for KYC verification endpoint (/api/v1/kyc/verify)."""
from unittest.mock import MagicMock, patch
from fastapi.testclient import TestClient


def test_kyc_verify_happy_path(client: TestClient):
    """Test successful KYC verification with mocked Groq vision response."""
    mock_completion = MagicMock()
    mock_completion.choices = [
        MagicMock(message=MagicMock(content='{"verification_score": 0.92, "reason": "ID matches selfie"}'))
    ]
    
    with patch("main._groq") as mock_groq:
        mock_groq.chat.completions.create.return_value = mock_completion
        
        response = client.post(
            "/api/v1/kyc/verify",
            json={
                "user_id": "usr_kyc_1",
                "id_card_url": "https://example.com/ktp.jpg",
                "selfie_url": "https://example.com/selfie.jpg",
                "target_role": "VERIFIED_MERCHANT"
            }
        )
        
        assert response.status_code == 200
        data = response.json()
        assert "score" in data
        assert data["score"] == 0.92
        assert data["reason"] == "ID matches selfie"


def test_kyc_verify_invalid_input(client: TestClient):
    """Test KYC verification with missing required fields expecting 422."""
    response = client.post(
        "/api/v1/kyc/verify",
        json={
            "user_id": "usr_kyc_1"
            # missing id_card_url and selfie_url
        }
    )
    assert response.status_code == 422


def test_kyc_verify_groq_error_returns_503(client: TestClient):
    """Test that Groq vision API exceptions return HTTP 503."""
    with patch("main._groq") as mock_groq:
        mock_groq.chat.completions.create.side_effect = Exception("Vision API timeout")
        
        response = client.post(
            "/api/v1/kyc/verify",
            json={
                "user_id": "usr_kyc_1",
                "id_card_url": "https://example.com/ktp.jpg",
                "selfie_url": "https://example.com/selfie.jpg"
            }
        )
        
        assert response.status_code == 503
        assert "verification-error" in response.json()["detail"]
