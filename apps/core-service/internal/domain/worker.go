package domain

import "context"

type CRMWorker interface {
	Start(ctx context.Context)
	Stop()
	ExecuteMonthlyEvaluation(ctx context.Context) error
}