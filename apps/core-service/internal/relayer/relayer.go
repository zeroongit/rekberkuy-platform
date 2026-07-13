package relayer

import "context"

// RelayerContract mendefinisikan blueprint untuk gasless transaction log ke Avalanche Fuji Testnet
type RelayerContract interface {
	LogTransactionOnChain(ctx context.Context, txID string, amount int64, buyer string, seller string) (string, error)
}

type relayerStub struct{}

func NewRelayerStub() RelayerContract {
	return &relayerStub{}
}

func (r *relayerStub) LogTransactionOnChain(ctx context.Context, txID string, amount int64, buyer string, seller string) (string, error) {
	// Stub: Nanti diisi implementasi go-ethereum client pada Minggu 4
	return "0xstubbedblockchaintxhash1234567890abcdef", nil
}