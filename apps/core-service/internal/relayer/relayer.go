package relayer

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"rekberkuy/core-service/internal/domain"
)

// This package is an ADAPTER that implements the domain.Relayer port.
// It records transaction audit-logs to Avalanche in a gasless way: the backend
// bears the gas fee (deployer wallet), users do not need a crypto wallet.
//
// Only HASH (txIdHash, buyerHash, sellerHash) + amount are stored on-chain.
// No sensitive data / PII, per the smart contract restrictions in CLAUDE.md.

// loggerABI defines the on-chain TransactionLogger contract interface.
// The targeted Solidity contract must provide the following function & event:
//
//	event TransactionLogged(bytes32 indexed txIdHash, uint256 amount, bytes32 indexed buyerHash, bytes32 indexed sellerHash, uint256 timestamp);
//	function logTransaction(bytes32 txIdHash, uint256 amount, bytes32 buyerHash, bytes32 sellerHash) external;
const loggerABI = `[
  {"anonymous":false,"inputs":[{"indexed":true,"internalType":"bytes32","name":"txIdHash","type":"bytes32"},{"indexed":false,"internalType":"uint256","name":"amount","type":"uint256"},{"indexed":true,"internalType":"bytes32","name":"buyerHash","type":"bytes32"},{"indexed":true,"internalType":"bytes32","name":"sellerHash","type":"bytes32"},{"indexed":false,"internalType":"uint256","name":"timestamp","type":"uint256"}],"name":"TransactionLogged","type":"event"},
  {"inputs":[{"internalType":"bytes32","name":"txIdHash","type":"bytes32"},{"internalType":"uint256","name":"amount","type":"uint256"},{"internalType":"bytes32","name":"buyerHash","type":"bytes32"},{"internalType":"bytes32","name":"sellerHash","type":"bytes32"}],"name":"logTransaction","outputs":[],"stateMutability":"nonpayable","type":"function"}
]`

const (
	loggerMethod  = "logTransaction"
	defaultGasLim = uint64(150_000) // safe margin for a single simple log function
	waitTimeout   = 60 * time.Second
	gasMargin     = 2 // fee cap = base fee * gasMargin + tip
)

// ethRelayer is the implementation of domain.Relayer via go-ethereum to Avalanche.
type ethRelayer struct {
	client       *ethclient.Client
	privateKey   *ecdsa.PrivateKey
	chainID      *big.Int
	contractAddr common.Address
	parsedABI    abi.ABI
}

// NewEthRelayer connects to the Avalanche node, loads the deployer private key,
// and prepares the ABI. Returns an error if any configuration fails
// (the caller may fall back to NewRelayerStub).
func NewEthRelayer(rpcURL, privateKeyHex, contractAddrHex string, chainID int64) (domain.Relayer, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to avalanche node: %w", err)
	}

	if chainID == 0 {
		chainIDBig, err := client.ChainID(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to get chain id from node: %w", err)
		}
		chainID = chainIDBig.Int64()
	}

	priv, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, fmt.Errorf("invalid deployer private key: %w", err)
	}

	parsed, err := abi.JSON(strings.NewReader(loggerABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse TransactionLogger ABI: %w", err)
	}

	if !common.IsHexAddress(contractAddrHex) {
		return nil, fmt.Errorf("invalid contract address: %s", contractAddrHex)
	}

	return &ethRelayer{
		client:       client,
		privateKey:   priv,
		chainID:      big.NewInt(chainID),
		contractAddr: common.HexToAddress(contractAddrHex),
		parsedABI:    parsed,
	}, nil
}

// LogTransactionOnChain records the transaction audit-log (hash + amount) to
// Avalanche and returns the transaction hash to be stored in the DB.
func (r *ethRelayer) LogTransactionOnChain(ctx context.Context, txID string, amount int64, buyer string, seller string) (string, error) {
	txIDHash := crypto.Keccak256Hash([]byte(txID))
	buyerHash := crypto.Keccak256Hash([]byte(buyer))
	sellerHash := crypto.Keccak256Hash([]byte(seller))
	amountBig := big.NewInt(amount)

	input, err := r.parsedABI.Pack(loggerMethod, txIDHash, amountBig, buyerHash, sellerHash)
	if err != nil {
		return "", fmt.Errorf("failed to pack logTransaction ABI: %w", err)
	}

	fromAddr := crypto.PubkeyToAddress(r.privateKey.PublicKey)
	nonce, err := r.client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	head, err := r.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get block header: %w", err)
	}
	baseFee := head.BaseFee
	if baseFee == nil {
		baseFee = big.NewInt(0)
	}

	tipCap, err := r.client.SuggestGasTipCap(ctx)
	if err != nil {
		tipCap = big.NewInt(1_000_000_000) // 1 gwei fallback
	}
	feeCap := new(big.Int).Mul(baseFee, big.NewInt(gasMargin))
	feeCap.Add(feeCap, tipCap)

	gasLimit := defaultGasLim
	if est, err := r.client.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddr, To: &r.contractAddr, Data: input,
	}); err == nil {
		gasLimit = est + est/5 // +20% margin
	}

	rawTx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   r.chainID,
		Nonce:     nonce,
		To:        &r.contractAddr,
		Value:     big.NewInt(0),
		Gas:       gasLimit,
		GasFeeCap: feeCap,
		GasTipCap: tipCap,
		Data:      input,
	})
	signedTx, err := types.SignTx(rawTx, types.NewLondonSigner(r.chainID), r.privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %w", err)
	}
	if err := r.client.SendTransaction(ctx, signedTx); err != nil {
		return "", fmt.Errorf("failed to send tx to avalanche: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()
	if _, err := bind.WaitMined(waitCtx, r.client, signedTx); err != nil {
		// tx sent but receipt not received within timeout -> still return the
		// hash so it can be polled; gas has already been paid.
		return signedTx.Hash().Hex(), fmt.Errorf("tx sent %s but receipt timeout: %w", signedTx.Hash().Hex(), err)
	}
	return signedTx.Hash().Hex(), nil
}

// --- Stub (used when node/contract is not configured) ---

type relayerStub struct{}

// NewRelayerStub returns a Relayer that returns a dummy tx hash.
// Used when the Avalanche node / contract is not configured (dev/test env).
func NewRelayerStub() domain.Relayer {
	return &relayerStub{}
}

func (r *relayerStub) LogTransactionOnChain(ctx context.Context, txID string, amount int64, buyer string, seller string) (string, error) {
	return "0xstubbedblockchaintxhash1234567890abcdef", nil
}
