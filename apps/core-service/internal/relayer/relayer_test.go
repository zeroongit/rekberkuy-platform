package relayer

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// TestLoggerABI_Parses ensures the ABI constant is valid & the logTransaction function exists.
func TestLoggerABI_Parses(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(loggerABI))
	if err != nil {
		t.Fatalf("failed to parse loggerABI: %v", err)
	}
	m, ok := parsed.Methods[loggerMethod]
	if !ok {
		t.Fatalf("method %s not found in ABI", loggerMethod)
	}
	if len(m.Inputs) != 4 {
		t.Errorf("logTransaction should have 4 inputs, got %d", len(m.Inputs))
	}
}

// TestLogCallCalldata_Deterministic verifies that calldata packing is consistent &
// deterministic for the same input (4-byte selector + 4 args of 32 bytes each).
func TestLogCallCalldata_Deterministic(t *testing.T) {
	parsed, _ := abi.JSON(strings.NewReader(loggerABI))

	txIDHash := crypto.Keccak256Hash([]byte("tx-123"))
	buyerHash := crypto.Keccak256Hash([]byte("buyer-1"))
	sellerHash := crypto.Keccak256Hash([]byte("seller-1"))
	amount := big.NewInt(500000)

	call1, err := parsed.Pack(loggerMethod, txIDHash, amount, buyerHash, sellerHash)
	if err != nil {
		t.Fatalf("pack failed: %v", err)
	}
	call2, _ := parsed.Pack(loggerMethod, txIDHash, amount, buyerHash, sellerHash)

	// 4-byte selector + 4 args of 32 bytes each = 4 + 128 = 132 bytes
	if len(call1) != 4+4*32 {
		t.Errorf("calldata length = %d, want %d", len(call1), 4+4*32)
	}
	sel := common.Bytes2Hex(call1[:4])
	if sel == "" {
		t.Error("selector must not be empty")
	}
	if common.Bytes2Hex(call1) != common.Bytes2Hex(call2) {
		t.Error("calldata should be identical for the same input")
	}
}

// TestHashing_Deterministic ensures the audit hashing (keccak) is stable for identical input.
func TestHashing_Deterministic(t *testing.T) {
	h1 := crypto.Keccak256Hash([]byte("tx-abc"))
	h2 := crypto.Keccak256Hash([]byte("tx-abc"))
	if h1 != h2 {
		t.Error("keccak hash should be deterministic")
	}
	if h1 == (common.Hash{}) {
		t.Error("hash must not be zero")
	}
}
