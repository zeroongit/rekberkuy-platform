package repository

import (
	"context"
	"database/sql"
	"fmt"
	"rekberkuy/core-service/internal/domain"
)

type kycRepository struct {
	db *sql.DB
}

func NewKYCRepository(db *sql.DB) domain.KYCRepository {
	return &kycRepository{db: db}
}

func (r *kycRepository) SubmitKYC(ctx context.Context, kyc *domain.KYCSubmission) error {
	query := `
		INSERT INTO kyc_submissions (id, user_id, target_role, id_card_number, id_card_url, selfie_url, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		ON CONFLICT (user_id) DO UPDATE SET 
			target_role = EXCLUDED.target_role,
			id_card_number = EXCLUDED.id_card_number, 
			id_card_url = EXCLUDED.id_card_url,
			selfie_url = EXCLUDED.selfie_url,
			status = 'PENDING',
			updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, 
		kyc.ID, 
		kyc.UserID, 
		string(kyc.TargetRole), 
		kyc.IDCardNumber, 
		kyc.IDCardURL, 
		kyc.SelfieURL, 
		string(kyc.Status),
	)
	if err != nil {
		return fmt.Errorf("gagal mencatat berkas kyc ke database: %w", err)
	}
	return nil
}