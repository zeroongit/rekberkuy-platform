package domain

import "context"

// CRMWorker mendefinisikan kontrak kerja untuk background engine otomatisasi
type CRMWorker interface {
	Start(ctx context.Context)
	Stop()
	ExecuteMonthlyEvaluation(ctx context.Context) error
}