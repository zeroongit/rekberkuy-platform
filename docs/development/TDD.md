# Test-Driven Development (TDD)
# RekberKuy Platform

> TDD adalah implementasi teknis dari BDD.
> Setiap scenario di BDD.md harus punya test di sini.
> Baca BDD.md terlebih dahulu.

---

## Siklus TDD — Red → Green → Refactor

```
1. RED    → Tulis test yang GAGAL dulu (fitur belum ada)
2. GREEN  → Tulis kode MINIMAL agar test lulus
3. REFACTOR → Bersihkan kode tanpa ubah behavior

JANGAN tulis kode sebelum ada test!
Kecuali untuk: main.go wiring, config loading, Dockerfile
```

---

## Piramida Test RekberKuy

```
           /▲\
          / E2E \           ← Playwright (e2e-qa/) — sedikit, lambat, mahal
         /───────\
        /Integration\       ← Postman / httptest Go — endpoint level
       /─────────────\
      /   Unit Tests   \    ← Jest / go test / pytest — banyak, cepat, murah
     /─────────────────────\
```

**Target coverage:**

| Layer | Minimum | Tool |
|-------|---------|------|
| `usecase/` Go | 80% | `go test -cover` |
| `repository/` Go | 70% | `go test -cover` |
| `delivery/handlers/` Go | 60% | `go test -cover` |
| `lib/utils/` Next.js | 90% | Jest |
| `components/` Next.js | 70% | Jest + RTL |
| `backend-ai` endpoints | 70% | pytest |

---

## TDD — Go (Core Service)

### Aturan

```
1. Setiap usecase baru → wajib ada *_usecase_test.go
2. Setiap repository baru → wajib ada *_repository_test.go
3. Test usecase menggunakan MOCK repository (bukan hit DB)
4. Test repository menggunakan DB test (testcontainer atau DB terpisah)
5. Gunakan table-driven test untuk multiple scenario
6. Jalankan dengan race detector: go test -race ./...
```

### Struktur Mock (Dependency Injection)

```go
// internal/domain/transaction.go — interface yang di-mock
type TransactionRepository interface {
    Create(ctx context.Context, tx *Transaction) error
    FindByID(ctx context.Context, id string) (*Transaction, error)
    UpdateStatus(ctx context.Context, id string, status TransactionStatus) error
}

// internal/usecase/mocks/transaction_repository_mock.go
type MockTransactionRepository struct {
    mock.Mock
}

func (m *MockTransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
    args := m.Called(ctx, tx)
    return args.Error(0)
}
```

### Template Unit Test Usecase

```go
// internal/usecase/transaction_goods_usecase_test.go
package usecase_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "rekberkuy/internal/domain"
    "rekberkuy/internal/usecase"
    "rekberkuy/internal/usecase/mocks"
)

func TestCreateGoodsTransaction(t *testing.T) {
    tests := []struct {
        name        string
        input       domain.CreateTransactionInput
        fraudScore  float64
        setupMock   func(*mocks.MockTransactionRepository, *mocks.MockWalletRepository, *mocks.MockFraudClient)
        expectError bool
        expectErrCode string
    }{
        {
            name: "success — fraud score rendah, escrow terkunci",
            input: domain.CreateTransactionInput{
                BuyerID:  "buyer-123",
                SellerID: "seller-456",
                Amount:   500000,
            },
            fraudScore: 0.2, // < threshold 0.5
            setupMock: func(txRepo *mocks.MockTransactionRepository, walletRepo *mocks.MockWalletRepository, fraud *mocks.MockFraudClient) {
                fraud.On("CheckRisk", mock.Anything, mock.Anything).Return(domain.FraudScore{Score: 0.2}, nil)
                walletRepo.On("LockFunds", mock.Anything, "buyer-123", int64(500000)).Return(nil)
                txRepo.On("Create", mock.Anything, mock.MatchedBy(func(tx *domain.Transaction) bool {
                    return tx.Status == domain.StatusActive
                })).Return(nil)
            },
            expectError: false,
        },
        {
            name: "gagal — fraud score tinggi, escrow diblokir",
            input: domain.CreateTransactionInput{
                BuyerID: "buyer-new",
                Amount:  50000000,
            },
            fraudScore: 0.85, // >= threshold 0.5
            setupMock: func(txRepo *mocks.MockTransactionRepository, walletRepo *mocks.MockWalletRepository, fraud *mocks.MockFraudClient) {
                fraud.On("CheckRisk", mock.Anything, mock.Anything).Return(domain.FraudScore{Score: 0.85}, nil)
                // walletRepo & txRepo TIDAK boleh dipanggil
            },
            expectError:   true,
            expectErrCode: "FRAUD_RISK_TOO_HIGH",
        },
        {
            name: "gagal — saldo tidak cukup",
            input: domain.CreateTransactionInput{
                BuyerID: "buyer-broke",
                Amount:  500000,
            },
            fraudScore: 0.1,
            setupMock: func(txRepo *mocks.MockTransactionRepository, walletRepo *mocks.MockWalletRepository, fraud *mocks.MockFraudClient) {
                fraud.On("CheckRisk", mock.Anything, mock.Anything).Return(domain.FraudScore{Score: 0.1}, nil)
                walletRepo.On("LockFunds", mock.Anything, "buyer-broke", int64(500000)).
                    Return(domain.ErrInsufficientBalance)
            },
            expectError:   true,
            expectErrCode: "INSUFFICIENT_BALANCE",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            txRepo := new(mocks.MockTransactionRepository)
            walletRepo := new(mocks.MockWalletRepository)
            fraudClient := new(mocks.MockFraudClient)
            tt.setupMock(txRepo, walletRepo, fraudClient)

            uc := usecase.NewTransactionGoodsUsecase(txRepo, walletRepo, fraudClient)
            result, err := uc.Create(context.Background(), tt.input)

            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectErrCode)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                assert.Equal(t, domain.StatusActive, result.Status)
            }

            txRepo.AssertExpectations(t)
            walletRepo.AssertExpectations(t)
            fraudClient.AssertExpectations(t)
        })
    }
}
```

### Template Unit Test Wallet — Race Condition

```go
// internal/usecase/wallet_usecase_test.go

func TestWalletLockFunds_Concurrent(t *testing.T) {
    // Test bahwa concurrent lock tidak menyebabkan double-debit
    // Ini HARUS dijalankan dengan -race flag

    var wg sync.WaitGroup
    errors := make([]error, 10)

    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            _, err := walletUsecase.LockFunds(ctx, "buyer-123", 100000)
            errors[idx] = err
        }(i)
    }

    wg.Wait()

    // Hanya 1 yang boleh sukses jika saldo hanya cukup untuk 1 transaksi
    successCount := 0
    for _, err := range errors {
        if err == nil {
            successCount++
        }
    }
    assert.Equal(t, 1, successCount, "Hanya 1 lock yang boleh berhasil")
}
```

### Jalankan Test

```bash
cd apps/core-service

# Semua test
go test ./...

# Dengan race detector (WAJIB untuk wallet & escrow)
go test -race ./...

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Test package spesifik
go test ./internal/usecase/... -v
go test ./internal/repository/... -v -run TestWallet

# Benchmark (opsional)
go test -bench=. ./internal/usecase/...
```

---

## TDD — Next.js (Dashboard Web)

### Aturan

```
1. Setiap komponen baru → wajib ada ComponentName.test.tsx
2. Setiap utility/hook baru → wajib ada filename.test.ts
3. Gunakan React Testing Library (RTL) — test behavior, bukan implementation
4. JANGAN test implementation detail (state internal, class name)
5. Test dari sudut pandang USER — apa yang mereka lihat dan klik
```

### Template Unit Test Komponen

```tsx
// components/modules/__tests__/TransactionCard.test.tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { TransactionCard } from '../TransactionCard'

const mockTransaction = {
  id: 'TRX-001',
  type: 'GOODS' as const,
  status: 'ACTIVE' as const,
  grossAmount: 500000,
  sellerName: 'Sari Penjual',
  createdAt: new Date('2026-01-01'),
}

describe('TransactionCard', () => {
  it('menampilkan nominal dalam format Rupiah', () => {
    render(<TransactionCard transaction={mockTransaction} />)
    expect(screen.getByText('Rp 500.000')).toBeInTheDocument()
  })

  it('menampilkan status ACTIVE dengan warna yang benar', () => {
    render(<TransactionCard transaction={mockTransaction} />)
    const badge = screen.getByText('Aktif')
    expect(badge).toHaveClass('bg-blue-100') // sesuaikan dengan class aktual
  })

  it('memanggil onConfirm dengan ID transaksi saat tombol diklik', async () => {
    const onConfirm = jest.fn()
    render(<TransactionCard transaction={mockTransaction} onConfirm={onConfirm} />)

    await userEvent.click(screen.getByRole('button', { name: /konfirmasi terima/i }))
    expect(onConfirm).toHaveBeenCalledWith('TRX-001')
    expect(onConfirm).toHaveBeenCalledTimes(1)
  })

  it('tidak menampilkan tombol konfirmasi jika status COMPLETED', () => {
    render(<TransactionCard transaction={{ ...mockTransaction, status: 'COMPLETED' }} />)
    expect(screen.queryByRole('button', { name: /konfirmasi/i })).not.toBeInTheDocument()
  })
})
```

### Template Unit Test Utility

```ts
// lib/utils/__tests__/formatCurrency.test.ts
import { formatRupiah } from '../formatCurrency'

describe('formatRupiah', () => {
  it.each([
    [0,         'Rp 0'],
    [500000,    'Rp 500.000'],
    [1500000,   'Rp 1.500.000'],
    [50000000,  'Rp 50.000.000'],
    [-50000,    '-Rp 50.000'],
  ])('formatRupiah(%i) → %s', (input, expected) => {
    expect(formatRupiah(input)).toBe(expected)
  })
})
```

### Jalankan Test

```bash
cd apps/dashboard-web

# Semua test
npm run test

# Watch mode (development)
npm run test -- --watch

# Coverage
npm run test -- --coverage

# Test file spesifik
npm run test -- TransactionCard.test.tsx

# Update snapshot (setelah UI change yang disengaja)
npm run test -- --updateSnapshot
```

---

## TDD — Python (Backend AI)

### Aturan

```
1. Setiap endpoint baru → wajib ada test_<endpoint>.py
2. Mock Groq API — JANGAN hit API asli saat test
3. Test mencakup: happy path, Groq unavailable (503), input invalid (422)
4. Gunakan FastAPI TestClient
```

### Template Unit Test Endpoint

```python
# backend-ai/tests/test_fraud.py
import pytest
from unittest.mock import patch, MagicMock
from fastapi.testclient import TestClient
from main import app

client = TestClient(app)

@pytest.fixture
def mock_groq_low_risk():
    """Mock Groq mengembalikan score rendah (transaksi normal)."""
    with patch("main._groq") as mock_groq:
        completion = MagicMock()
        completion.choices[0].message.content = '{"risk_score": 0.15, "reason": "normal transaction"}'
        mock_groq.chat.completions.create.return_value = completion
        yield mock_groq

@pytest.fixture
def mock_groq_high_risk():
    """Mock Groq mengembalikan score tinggi (transaksi mencurigakan)."""
    with patch("main._groq") as mock_groq:
        completion = MagicMock()
        completion.choices[0].message.content = '{"risk_score": 0.92, "reason": "new user high amount"}'
        mock_groq.chat.completions.create.return_value = completion
        yield mock_groq

class TestFraudScore:
    def test_score_normal_transaction(self, mock_groq_low_risk):
        """Transaksi normal dari user established → score rendah."""
        response = client.post("/api/v1/fraud/score", json={
            "user_id": "user-established-123",
            "amount": 150000
        })
        assert response.status_code == 200
        data = response.json()
        assert "score" in data
        assert 0.0 <= data["score"] <= 1.0
        assert data["score"] < 0.5

    def test_score_suspicious_transaction(self, mock_groq_high_risk):
        """Transaksi besar dari user baru → score tinggi."""
        response = client.post("/api/v1/fraud/score", json={
            "user_id": "user-new-001",
            "amount": 50000000
        })
        assert response.status_code == 200
        assert response.json()["score"] >= 0.5

    def test_missing_required_field(self):
        """Request tanpa amount → 422 Validation Error."""
        response = client.post("/api/v1/fraud/score", json={
            "user_id": "user-123"
            # amount missing
        })
        assert response.status_code == 422

    def test_groq_unavailable_returns_503(self):
        """Jika Groq error → 503 (core-service applies fail-open/closed)."""
        with patch("main._groq") as mock_groq:
            mock_groq.chat.completions.create.side_effect = Exception("Groq timeout")
            response = client.post("/api/v1/fraud/score", json={
                "user_id": "user-123",
                "amount": 100000
            })
        assert response.status_code == 503

    def test_no_groq_key_returns_fallback(self):
        """Tanpa GROQ_API_KEY → fallback score 0.1 (development mode)."""
        with patch("main._groq", None):
            response = client.post("/api/v1/fraud/score", json={
                "user_id": "user-123",
                "amount": 100000
            })
        assert response.status_code == 200
        data = response.json()
        assert data["score"] == 0.1
        assert "fallback" in data["reason"]

class TestKYCVerify:
    def test_verify_valid_documents(self):
        """KTP + selfie valid → score tinggi."""
        with patch("main._groq") as mock_groq:
            completion = MagicMock()
            completion.choices[0].message.content = '{"verification_score": 0.92, "reason": "documents match"}'
            mock_groq.chat.completions.create.return_value = completion

            response = client.post("/api/v1/kyc/verify", json={
                "user_id": "user-123",
                "id_card_url": "https://storage.supabase.co/ktp-123.jpg",
                "selfie_url": "https://storage.supabase.co/selfie-123.jpg",
                "target_role": "VERIFIED_MERCHANT"
            })
        assert response.status_code == 200
        assert response.json()["score"] >= 0.8

    def test_healthz_endpoint(self):
        """Health check selalu 200."""
        response = client.get("/healthz")
        assert response.status_code == 200
        assert response.json()["status"] == "ok"
```

### Jalankan Test

```bash
cd backend-ai

# Install test dependencies
pip install pytest pytest-asyncio httpx

# Jalankan semua test
pytest

# Verbose output
pytest -v

# Coverage
pytest --cov=. --cov-report=html

# Test file spesifik
pytest tests/test_fraud.py -v
pytest tests/test_fraud.py::TestFraudScore::test_groq_unavailable_returns_503
```

---

## Checklist TDD Sebelum Merge PR

```
Go (Core Service):
□ go test -race ./... → semua hijau
□ Coverage usecase >= 80%
□ Coverage repository >= 70%
□ Tidak ada test yang di-skip dengan t.Skip() tanpa alasan

Next.js (Dashboard):
□ npm run test → semua hijau
□ Tidak ada console.error di test output
□ Tidak ada act() warning (update ke await userEvent)

Python (Backend AI):
□ pytest → semua hijau
□ Semua Groq call di-mock (tidak hit API asli)
□ Test mencakup: happy path, Groq unavailable, input invalid
```

---

## Anti-Pattern yang HARUS Dihindari

```go
// ❌ Test yang tidak berguna — hanya test bahwa function terpanggil
func TestCreateTransaction_Bad(t *testing.T) {
    mock.On("Create", ...).Return(nil)
    uc.Create(ctx, input)
    mock.AssertCalled(t, "Create") // ini tidak test behavior apapun
}

// ✅ Test yang berguna — test outcome bisnis
func TestCreateTransaction_Good(t *testing.T) {
    result, err := uc.Create(ctx, validInput)
    assert.NoError(t, err)
    assert.Equal(t, domain.StatusActive, result.Status)    // ← test STATE
    assert.Equal(t, int64(492500), result.NetPayout)       // ← test KALKULASI
    walletMock.AssertCalled(t, "LockFunds", "buyer-123", int64(500000)) // ← test SIDE EFFECT
}
```

```tsx
// ❌ Test implementation detail
expect(component.state.isLoading).toBe(false)

// ✅ Test dari sudut pandang user
expect(screen.queryByTestId('loading-spinner')).not.toBeInTheDocument()
expect(screen.getByText('Rp 500.000')).toBeInTheDocument()
```