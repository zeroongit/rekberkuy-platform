package fraud

import "context"

// FraudClientContract mendefinisikan jembatan komunikasi ke microservice Python FastAPI
type FraudClientContract interface {
	AnalyzeTransactionRisk(ctx context.Context, userID string, amount int64) (float64, bool, error)
}

type fraudClientStub struct{}

func NewFraudClientStub() FraudClientContract {
	return &fraudClientStub{}
}

func (f *fraudClientStub) AnalyzeTransactionRisk(ctx context.Context, userID string, amount int64) (float64, bool, error) {
	// Stub: Nanti diisi REST/gRPC call ke Python Minggu 5. Sementara return score aman (0.1) dan isSafe (true).
	return 0.1, true, nil
}