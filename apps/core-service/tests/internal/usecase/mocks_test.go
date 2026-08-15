package usecase_test

import (
	"context"
	"errors"
	"time"

	"rekberkuy/core-service/internal/domain"
)

// ============================================================================
// SHARED TEST FIXTURES — valid detail containers for lock-flow tests
// ============================================================================

func validGoodsDetail() *domain.TransactionGoods {
	return &domain.TransactionGoods{
		SubSubCategoryID:    1,
		ShippingCourier:     "JNE",
		ShippingAddress:     "Jl. Testing No. 1, Jakarta",
		AutoConfirmDeadline: time.Now().Add(48 * time.Hour),
	}
}

func validServicesDetail() *domain.TransactionServices {
	return &domain.TransactionServices{
		SubSubCategoryID:    1,
		ProjectDeadline:    time.Now().Add(7 * 24 * time.Hour),
		BriefDescription:   "Build a landing page",
	}
}

// validServicesMilestones returns two milestones summing exactly to total.
func validServicesMilestones(total int64) []domain.ServiceMilestone {
	return []domain.ServiceMilestone{
		{Title: "Kickoff & design", Amount: total / 2},
		{Title: "Final delivery", Amount: total - total/2},
	}
}

func validEventsDetail() *domain.TransactionEvents {
	return &domain.TransactionEvents{
		SubSubCategoryID:    1,
		EventName:           "Wedding RekberKuy",
		EventStartTime:      time.Now().Add(30 * 24 * time.Hour),
		EventEndTime:        time.Now().Add(30 * 24 * time.Hour).Add(8 * time.Hour),
		TicketQuantityTotal: 100,
	}
}

// ============================================================================
// SHARED MOCK REPOSITORIES
// ----------------------------------------------------------------------------
// Convention: each mock embeds the domain interface (so non-overridden methods
// panic if called — fail-fast). Commonly-used methods are overridden via function
// fields; if the field is nil, return a safe zero-value.
// This pattern follows the existing transaction_goods_usecase_test.go.
// ============================================================================

// ---------------------------------------------------------------------------
// TransactionRepository
// ---------------------------------------------------------------------------

type mockTransactionRepo struct {
	domain.TransactionRepository
	onCreateTransaction              func(ctx context.Context, tx *domain.Transaction) error
	onGetTransactionByID             func(ctx context.Context, id string) (*domain.Transaction, error)
	onUpdateTransactionStatus        func(ctx context.Context, id string, status domain.TransactionStatus) error
	onGetExpiredLocked               func(ctx context.Context) ([]string, error)
	onGetReleasedMilestonesTotal     func(ctx context.Context, txID string) (int64, error)
	onGetMilestoneByID               func(ctx context.Context, id string) (*domain.ServiceMilestone, error)
	onUpdateMilestoneStatus          func(ctx context.Context, id string, status string) error
	onGetEventVendorPayouts          func(ctx context.Context, txID string) ([]domain.EventVendorPayout, error)
	onGetEventVendorPayoutByID       func(ctx context.Context, payoutID string) (*domain.EventVendorPayout, error)
	onUpdateEventVendorPayoutStatus  func(ctx context.Context, id string, status string) error
	onMarkEventVendorPayoutDisbursed func(ctx context.Context, payoutID string, adminID string) error
	onUpdateBlockchainLog            func(ctx context.Context, txID string, txHash string) error
	onGetByMidtransOrderID           func(ctx context.Context, orderID string) (*domain.Transaction, error)
	onCreateGoodsDetail              func(ctx context.Context, tg *domain.TransactionGoods) error
	onCreateServicesDetail           func(ctx context.Context, ts *domain.TransactionServices, milestones []domain.ServiceMilestone) error
	onCreateEventsDetail             func(ctx context.Context, te *domain.TransactionEvents, allocations []domain.EventVendorAllocation) error
	onGetActivePayoutsTotal          func(ctx context.Context, txID string) (int64, error)
	onListTransactionsByUser         func(ctx context.Context, userID string, limit, offset int) ([]domain.Transaction, error)
	onGetGoodsDetail                 func(ctx context.Context, txID string) (*domain.TransactionGoods, error)
	onGetServicesDetail              func(ctx context.Context, txID string) (*domain.TransactionServices, error)
	onGetMilestones                  func(ctx context.Context, txID string) ([]domain.ServiceMilestone, error)
	onGetEventsDetail                func(ctx context.Context, txID string) (*domain.TransactionEvents, error)
}

func (m *mockTransactionRepo) CreateTransaction(ctx context.Context, tx *domain.Transaction) error {
	if m.onCreateTransaction != nil {
		return m.onCreateTransaction(ctx, tx)
	}
	return nil
}

func (m *mockTransactionRepo) GetTransactionByID(ctx context.Context, id string) (*domain.Transaction, error) {
	if m.onGetTransactionByID != nil {
		return m.onGetTransactionByID(ctx, id)
	}
	return &domain.Transaction{}, nil
}

func (m *mockTransactionRepo) UpdateTransactionStatus(ctx context.Context, id string, status domain.TransactionStatus) error {
	if m.onUpdateTransactionStatus != nil {
		return m.onUpdateTransactionStatus(ctx, id, status)
	}
	return nil
}

func (m *mockTransactionRepo) GetExpiredLockedTransactions(ctx context.Context) ([]string, error) {
	if m.onGetExpiredLocked != nil {
		return m.onGetExpiredLocked(ctx)
	}
	return nil, nil
}

func (m *mockTransactionRepo) GetReleasedMilestonesTotalByTxID(ctx context.Context, txID string) (int64, error) {
	if m.onGetReleasedMilestonesTotal != nil {
		return m.onGetReleasedMilestonesTotal(ctx, txID)
	}
	return 0, nil
}

func (m *mockTransactionRepo) GetMilestoneByID(ctx context.Context, id string) (*domain.ServiceMilestone, error) {
	if m.onGetMilestoneByID != nil {
		return m.onGetMilestoneByID(ctx, id)
	}
	return &domain.ServiceMilestone{}, nil
}

func (m *mockTransactionRepo) UpdateMilestoneStatus(ctx context.Context, id string, status string) error {
	if m.onUpdateMilestoneStatus != nil {
		return m.onUpdateMilestoneStatus(ctx, id, status)
	}
	return nil
}

func (m *mockTransactionRepo) GetEventVendorPayoutsByTxID(ctx context.Context, txID string) ([]domain.EventVendorPayout, error) {
	if m.onGetEventVendorPayouts != nil {
		return m.onGetEventVendorPayouts(ctx, txID)
	}
	return nil, nil
}

func (m *mockTransactionRepo) UpdateEventVendorPayoutStatus(ctx context.Context, id string, status string) error {
	if m.onUpdateEventVendorPayoutStatus != nil {
		return m.onUpdateEventVendorPayoutStatus(ctx, id, status)
	}
	return nil
}

func (m *mockTransactionRepo) GetEventVendorPayoutByID(ctx context.Context, payoutID string) (*domain.EventVendorPayout, error) {
	if m.onGetEventVendorPayoutByID != nil {
		return m.onGetEventVendorPayoutByID(ctx, payoutID)
	}
	return &domain.EventVendorPayout{}, nil
}

func (m *mockTransactionRepo) MarkEventVendorPayoutDisbursed(ctx context.Context, payoutID string, adminID string) error {
	if m.onMarkEventVendorPayoutDisbursed != nil {
		return m.onMarkEventVendorPayoutDisbursed(ctx, payoutID, adminID)
	}
	return nil
}

func (m *mockTransactionRepo) UpdateBlockchainLog(ctx context.Context, txID string, txHash string) error {
	if m.onUpdateBlockchainLog != nil {
		return m.onUpdateBlockchainLog(ctx, txID, txHash)
	}
	return nil
}

func (m *mockTransactionRepo) GetTransactionByMidtransOrderID(ctx context.Context, orderID string) (*domain.Transaction, error) {
	if m.onGetByMidtransOrderID != nil {
		return m.onGetByMidtransOrderID(ctx, orderID)
	}
	return &domain.Transaction{}, nil
}

func (m *mockTransactionRepo) CreateGoodsDetail(ctx context.Context, tg *domain.TransactionGoods) error {
	if m.onCreateGoodsDetail != nil {
		return m.onCreateGoodsDetail(ctx, tg)
	}
	return nil
}

func (m *mockTransactionRepo) CreateServicesDetail(ctx context.Context, ts *domain.TransactionServices, milestones []domain.ServiceMilestone) error {
	if m.onCreateServicesDetail != nil {
		return m.onCreateServicesDetail(ctx, ts, milestones)
	}
	return nil
}

func (m *mockTransactionRepo) CreateEventsDetail(ctx context.Context, te *domain.TransactionEvents, allocations []domain.EventVendorAllocation) error {
	if m.onCreateEventsDetail != nil {
		return m.onCreateEventsDetail(ctx, te, allocations)
	}
	return nil
}

func (m *mockTransactionRepo) GetActiveEventVendorPayoutsTotal(ctx context.Context, txID string) (int64, error) {
	if m.onGetActivePayoutsTotal != nil {
		return m.onGetActivePayoutsTotal(ctx, txID)
	}
	return 0, nil
}

func (m *mockTransactionRepo) ListTransactionsByUser(ctx context.Context, userID string, limit, offset int) ([]domain.Transaction, error) {
	if m.onListTransactionsByUser != nil {
		return m.onListTransactionsByUser(ctx, userID, limit, offset)
	}
	return nil, nil
}

func (m *mockTransactionRepo) GetGoodsDetailByTxID(ctx context.Context, txID string) (*domain.TransactionGoods, error) {
	if m.onGetGoodsDetail != nil {
		return m.onGetGoodsDetail(ctx, txID)
	}
	return nil, errors.New("not found")
}

func (m *mockTransactionRepo) GetServicesDetailByTxID(ctx context.Context, txID string) (*domain.TransactionServices, error) {
	if m.onGetServicesDetail != nil {
		return m.onGetServicesDetail(ctx, txID)
	}
	return nil, errors.New("not found")
}

func (m *mockTransactionRepo) GetMilestonesByTxID(ctx context.Context, txID string) ([]domain.ServiceMilestone, error) {
	if m.onGetMilestones != nil {
		return m.onGetMilestones(ctx, txID)
	}
	return nil, nil
}

func (m *mockTransactionRepo) GetEventsDetailByTxID(ctx context.Context, txID string) (*domain.TransactionEvents, error) {
	if m.onGetEventsDetail != nil {
		return m.onGetEventsDetail(ctx, txID)
	}
	return nil, errors.New("not found")
}

// ---------------------------------------------------------------------------
// WalletRepository
// ---------------------------------------------------------------------------

type mockWalletRepo struct {
	domain.WalletRepository
	onUpdateBalanceTx          func(ctx context.Context, txRecord *domain.RekberPayTransaction, modifier int64) error
	onGetBalance               func(ctx context.Context, userID string) (*domain.RekberPayWallet, error)
	onCreateWallet             func(ctx context.Context, userID string) error
	onGetAllUsersForCRM        func(ctx context.Context) ([]*domain.UserProfile, error)
	onGetCRMLoyaltyByUserID    func(ctx context.Context, userID string) (*domain.CRMLoyalty, error)
	onUpdateCRMLoyalty         func(ctx context.Context, crmProfile *domain.CRMLoyalty) error
	onGetVendorAllocations     func(ctx context.Context, transactionID string) ([]*domain.EventVendorAllocation, error)
	onMarkAllocClaimed         func(ctx context.Context, transactionID, vendorID string, amount int64) error
	onCreateVendorPayoutRecord func(ctx context.Context, payout *domain.EventVendorPayout) error
	onGetVendorPayoutByID      func(ctx context.Context, payoutID string) (*domain.EventVendorPayout, error)
	onUpdateVendorPayoutStatus func(ctx context.Context, payoutID string, status string) error
	onGetWalletTxByOrderID     func(ctx context.Context, orderID string) (*domain.RekberPayTransaction, error)
	onMarkWalletTxStatus       func(ctx context.Context, orderID string, status domain.WalletTxStatus) error
	onGetWalletTxHistory       func(ctx context.Context, userID string, limit, offset int) ([]domain.RekberPayTransaction, error)
	onCreateWithdrawal         func(ctx context.Context, w *domain.WithdrawalRequest) error
	onGetWithdrawalByID        func(ctx context.Context, id string) (*domain.WithdrawalRequest, error)
	onListWithdrawalsByUser    func(ctx context.Context, userID string, limit, offset int) ([]domain.WithdrawalRequest, error)
	onListPendingWithdrawals   func(ctx context.Context, limit, offset int) ([]domain.WithdrawalRequest, error)
	onMarkWithdrawalDisbursed  func(ctx context.Context, id string, adminID string, actualCost *int64) error
}

func (m *mockWalletRepo) UpdateBalanceTx(ctx context.Context, txRecord *domain.RekberPayTransaction, modifier int64) error {
	if m.onUpdateBalanceTx != nil {
		return m.onUpdateBalanceTx(ctx, txRecord, modifier)
	}
	return nil
}

func (m *mockWalletRepo) GetBalance(ctx context.Context, userID string) (*domain.RekberPayWallet, error) {
	if m.onGetBalance != nil {
		return m.onGetBalance(ctx, userID)
	}
	return &domain.RekberPayWallet{UserID: userID}, nil
}

func (m *mockWalletRepo) CreateWallet(ctx context.Context, userID string) error {
	if m.onCreateWallet != nil {
		return m.onCreateWallet(ctx, userID)
	}
	return nil
}

func (m *mockWalletRepo) GetAllUsersForCRMEvaluation(ctx context.Context) ([]*domain.UserProfile, error) {
	if m.onGetAllUsersForCRM != nil {
		return m.onGetAllUsersForCRM(ctx)
	}
	return nil, nil
}

func (m *mockWalletRepo) GetCRMLoyaltyByUserID(ctx context.Context, userID string) (*domain.CRMLoyalty, error) {
	if m.onGetCRMLoyaltyByUserID != nil {
		return m.onGetCRMLoyaltyByUserID(ctx, userID)
	}
	return &domain.CRMLoyalty{UserID: userID}, nil
}

func (m *mockWalletRepo) UpdateCRMLoyalty(ctx context.Context, crmProfile *domain.CRMLoyalty) error {
	if m.onUpdateCRMLoyalty != nil {
		return m.onUpdateCRMLoyalty(ctx, crmProfile)
	}
	return nil
}

func (m *mockWalletRepo) GetVendorAllocationsByTxID(ctx context.Context, transactionID string) ([]*domain.EventVendorAllocation, error) {
	if m.onGetVendorAllocations != nil {
		return m.onGetVendorAllocations(ctx, transactionID)
	}
	return nil, nil
}

func (m *mockWalletRepo) MarkVendorAllocationClaimed(ctx context.Context, transactionID, vendorID string, amount int64) error {
	if m.onMarkAllocClaimed != nil {
		return m.onMarkAllocClaimed(ctx, transactionID, vendorID, amount)
	}
	return nil
}

func (m *mockWalletRepo) CreateVendorPayoutRecord(ctx context.Context, payout *domain.EventVendorPayout) error {
	if m.onCreateVendorPayoutRecord != nil {
		return m.onCreateVendorPayoutRecord(ctx, payout)
	}
	return nil
}

func (m *mockWalletRepo) GetVendorPayoutByID(ctx context.Context, payoutID string) (*domain.EventVendorPayout, error) {
	if m.onGetVendorPayoutByID != nil {
		return m.onGetVendorPayoutByID(ctx, payoutID)
	}
	return &domain.EventVendorPayout{}, nil
}

func (m *mockWalletRepo) UpdateVendorPayoutStatus(ctx context.Context, payoutID string, status string) error {
	if m.onUpdateVendorPayoutStatus != nil {
		return m.onUpdateVendorPayoutStatus(ctx, payoutID, status)
	}
	return nil
}

func (m *mockWalletRepo) GetWalletTxByMidtransOrderID(ctx context.Context, orderID string) (*domain.RekberPayTransaction, error) {
	if m.onGetWalletTxByOrderID != nil {
		return m.onGetWalletTxByOrderID(ctx, orderID)
	}
	return &domain.RekberPayTransaction{MidtransTopUpID: &orderID}, nil
}

func (m *mockWalletRepo) MarkWalletTxStatusByOrderID(ctx context.Context, orderID string, status domain.WalletTxStatus) error {
	if m.onMarkWalletTxStatus != nil {
		return m.onMarkWalletTxStatus(ctx, orderID, status)
	}
	return nil
}

func (m *mockWalletRepo) GetWalletTxHistory(ctx context.Context, userID string, limit, offset int) ([]domain.RekberPayTransaction, error) {
	if m.onGetWalletTxHistory != nil {
		return m.onGetWalletTxHistory(ctx, userID, limit, offset)
	}
	return nil, nil
}

func (m *mockWalletRepo) CreateWithdrawalRequest(ctx context.Context, w *domain.WithdrawalRequest) error {
	if m.onCreateWithdrawal != nil {
		return m.onCreateWithdrawal(ctx, w)
	}
	return nil
}

func (m *mockWalletRepo) GetWithdrawalByID(ctx context.Context, id string) (*domain.WithdrawalRequest, error) {
	if m.onGetWithdrawalByID != nil {
		return m.onGetWithdrawalByID(ctx, id)
	}
	return &domain.WithdrawalRequest{ID: id, Status: domain.WithdrawalPending}, nil
}

func (m *mockWalletRepo) ListWithdrawalsByUser(ctx context.Context, userID string, limit, offset int) ([]domain.WithdrawalRequest, error) {
	if m.onListWithdrawalsByUser != nil {
		return m.onListWithdrawalsByUser(ctx, userID, limit, offset)
	}
	return nil, nil
}

func (m *mockWalletRepo) ListPendingWithdrawals(ctx context.Context, limit, offset int) ([]domain.WithdrawalRequest, error) {
	if m.onListPendingWithdrawals != nil {
		return m.onListPendingWithdrawals(ctx, limit, offset)
	}
	return nil, nil
}

func (m *mockWalletRepo) MarkWithdrawalDisbursed(ctx context.Context, id string, adminID string, actualCost *int64) error {
	if m.onMarkWithdrawalDisbursed != nil {
		return m.onMarkWithdrawalDisbursed(ctx, id, adminID, actualCost)
	}
	return nil
}

// ---------------------------------------------------------------------------
// DisbursementClient
// ---------------------------------------------------------------------------

// mockDisbursementClient default: flat estimate, execution refused — mirrors
// the production stub. Override the fields for specific cases.
type mockDisbursementClient struct {
	domain.DisbursementClient
	estimate int64
	onEstimate func(ctx context.Context, req domain.DisbursementRequest) (int64, error)
}

func (m *mockDisbursementClient) EstimateDisbursementFee(ctx context.Context, req domain.DisbursementRequest) (int64, error) {
	if m.onEstimate != nil {
		return m.onEstimate(ctx, req)
	}
	if m.estimate != 0 {
		return m.estimate, nil
	}
	return domain.WithdrawFeeMidtransCost, nil
}

func (m *mockDisbursementClient) ExecuteDisbursement(ctx context.Context, req domain.DisbursementRequest) (domain.DisbursementResult, error) {
	return domain.DisbursementResult{}, errors.New("not wired")
}

// ---------------------------------------------------------------------------
// FinanceRepository
// ---------------------------------------------------------------------------

type mockFinanceRepo struct {
	domain.FinanceRepository
	onUpdatePlatformFinance func(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error
	onGetPlatformFinance    func(ctx context.Context) (*domain.PlatformFinance, error)
}

func (m *mockFinanceRepo) UpdatePlatformFinance(ctx context.Context, escrowDelta, revenueDelta, midtransFeeDelta int64) error {
	if m.onUpdatePlatformFinance != nil {
		return m.onUpdatePlatformFinance(ctx, escrowDelta, revenueDelta, midtransFeeDelta)
	}
	return nil
}

func (m *mockFinanceRepo) GetPlatformFinance(ctx context.Context) (*domain.PlatformFinance, error) {
	if m.onGetPlatformFinance != nil {
		return m.onGetPlatformFinance(ctx)
	}
	return &domain.PlatformFinance{}, nil
}

// ---------------------------------------------------------------------------
// UserRepository
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	domain.UserRepository
	onCreateProfile     func(ctx context.Context, user *domain.UserProfile) error
	onGetProfileByID    func(ctx context.Context, id string) (*domain.UserProfile, error)
	onGetProfileByEmail func(ctx context.Context, email string) (*domain.UserProfile, error)
	onUpdateUserRole    func(ctx context.Context, id string, role domain.UserRole) error
}

func (m *mockUserRepo) CreateProfile(ctx context.Context, user *domain.UserProfile) error {
	if m.onCreateProfile != nil {
		return m.onCreateProfile(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetProfileByID(ctx context.Context, id string) (*domain.UserProfile, error) {
	if m.onGetProfileByID != nil {
		return m.onGetProfileByID(ctx, id)
	}
	return &domain.UserProfile{ID: id}, nil
}

func (m *mockUserRepo) GetProfileByEmail(ctx context.Context, email string) (*domain.UserProfile, error) {
	if m.onGetProfileByEmail != nil {
		return m.onGetProfileByEmail(ctx, email)
	}
	return nil, errors.New("not found")
}

func (m *mockUserRepo) UpdateUserRole(ctx context.Context, id string, role domain.UserRole) error {
	if m.onUpdateUserRole != nil {
		return m.onUpdateUserRole(ctx, id, role)
	}
	return nil
}

// ---------------------------------------------------------------------------
// KYCRepository
// ---------------------------------------------------------------------------

type mockKYCRepo struct {
	domain.KYCRepository
	onSubmitKYC       func(ctx context.Context, kyc *domain.KYCSubmission) error
	onGetKYCByID      func(ctx context.Context, id string) (*domain.KYCSubmission, error)
	onListPendingKYCs func(ctx context.Context) ([]domain.KYCSubmission, error)
	onSaveKYCAIResult func(ctx context.Context, userID string, score float64, reason string) error
	onReviewKYC       func(ctx context.Context, id string, status domain.KYCStatus, adminID string, notes string) error
}

func (m *mockKYCRepo) SubmitKYC(ctx context.Context, kyc *domain.KYCSubmission) error {
	if m.onSubmitKYC != nil {
		return m.onSubmitKYC(ctx, kyc)
	}
	return nil
}

func (m *mockKYCRepo) GetKYCByID(ctx context.Context, id string) (*domain.KYCSubmission, error) {
	if m.onGetKYCByID != nil {
		return m.onGetKYCByID(ctx, id)
	}
	return &domain.KYCSubmission{ID: id, Status: domain.KYCPending}, nil
}

func (m *mockKYCRepo) ListPendingKYCs(ctx context.Context) ([]domain.KYCSubmission, error) {
	if m.onListPendingKYCs != nil {
		return m.onListPendingKYCs(ctx)
	}
	return nil, nil
}

func (m *mockKYCRepo) SaveKYCAIResult(ctx context.Context, userID string, score float64, reason string) error {
	if m.onSaveKYCAIResult != nil {
		return m.onSaveKYCAIResult(ctx, userID, score, reason)
	}
	return nil
}

func (m *mockKYCRepo) ReviewKYC(ctx context.Context, id string, status domain.KYCStatus, adminID string, notes string) error {
	if m.onReviewKYC != nil {
		return m.onReviewKYC(ctx, id, status, adminID, notes)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ReviewRepository
// ---------------------------------------------------------------------------

type mockReviewRepo struct {
	domain.ReviewRepository
	onCreateReview            func(ctx context.Context, review *domain.Review) error
	onGetReviewsByReviewee    func(ctx context.Context, revieweeID string) ([]domain.Review, error)
	onGetAverageRatingForUser func(ctx context.Context, userID string) (float64, error)
}

func (m *mockReviewRepo) CreateReview(ctx context.Context, review *domain.Review) error {
	if m.onCreateReview != nil {
		return m.onCreateReview(ctx, review)
	}
	return nil
}

func (m *mockReviewRepo) GetReviewsByReviewee(ctx context.Context, revieweeID string) ([]domain.Review, error) {
	if m.onGetReviewsByReviewee != nil {
		return m.onGetReviewsByReviewee(ctx, revieweeID)
	}
	return nil, nil
}

func (m *mockReviewRepo) GetAverageRatingForUser(ctx context.Context, userID string) (float64, error) {
	if m.onGetAverageRatingForUser != nil {
		return m.onGetAverageRatingForUser(ctx, userID)
	}
	return 0, nil
}

// ---------------------------------------------------------------------------
// DisputeRepository
// ---------------------------------------------------------------------------

type mockDisputeRepo struct {
	domain.DisputeRepository
	onCreateDispute             func(ctx context.Context, d *domain.Dispute) error
	onGetDisputeByID            func(ctx context.Context, id string) (*domain.Dispute, error)
	onGetDisputeByTransactionID func(ctx context.Context, transactionID string) (*domain.Dispute, error)
	onAcknowledgeDispute        func(ctx context.Context, id string) error
	onResolveDispute            func(ctx context.Context, id string, adminID string, outcome domain.DisputeOutcome, summary string) error
}

func (m *mockDisputeRepo) CreateDispute(ctx context.Context, d *domain.Dispute) error {
	if m.onCreateDispute != nil {
		return m.onCreateDispute(ctx, d)
	}
	return nil
}

func (m *mockDisputeRepo) GetDisputeByID(ctx context.Context, id string) (*domain.Dispute, error) {
	if m.onGetDisputeByID != nil {
		return m.onGetDisputeByID(ctx, id)
	}
	return &domain.Dispute{ID: id}, nil
}

func (m *mockDisputeRepo) GetDisputeByTransactionID(ctx context.Context, transactionID string) (*domain.Dispute, error) {
	if m.onGetDisputeByTransactionID != nil {
		return m.onGetDisputeByTransactionID(ctx, transactionID)
	}
	return nil, errors.New("not found")
}

func (m *mockDisputeRepo) AcknowledgeDispute(ctx context.Context, id string) error {
	if m.onAcknowledgeDispute != nil {
		return m.onAcknowledgeDispute(ctx, id)
	}
	return nil
}

func (m *mockDisputeRepo) ResolveDispute(ctx context.Context, id string, adminID string, outcome domain.DisputeOutcome, summary string) error {
	if m.onResolveDispute != nil {
		return m.onResolveDispute(ctx, id, adminID, outcome, summary)
	}
	return nil
}

// ---------------------------------------------------------------------------
// VendorRepository
// ---------------------------------------------------------------------------

type mockVendorRepo struct {
	domain.VendorRepository
	onCreateVendor func(ctx context.Context, vendor *domain.VendorProfile) error
	onListVendors  func(ctx context.Context, category string, limit, offset int) ([]domain.VendorProfile, error)
}

func (m *mockVendorRepo) CreateVendor(ctx context.Context, vendor *domain.VendorProfile) error {
	if m.onCreateVendor != nil {
		return m.onCreateVendor(ctx, vendor)
	}
	return nil
}

func (m *mockVendorRepo) ListVendors(ctx context.Context, category string, limit, offset int) ([]domain.VendorProfile, error) {
	if m.onListVendors != nil {
		return m.onListVendors(ctx, category, limit, offset)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// CategoryRepository
// ---------------------------------------------------------------------------

type mockCategoryRepo struct {
	domain.CategoryRepository
	onListGoods   func(ctx context.Context) ([]domain.GoodsCategory, error)
	onListService func(ctx context.Context) ([]domain.ServiceCategory, error)
	onListEvent   func(ctx context.Context) ([]domain.EventCategory, error)
	onListVendor  func(ctx context.Context) ([]domain.VendorSubCategory, error)
}

func (m *mockCategoryRepo) ListGoodsCategories(ctx context.Context) ([]domain.GoodsCategory, error) {
	if m.onListGoods != nil {
		return m.onListGoods(ctx)
	}
	return []domain.GoodsCategory{{ID: 1, Name: "Electronics", Slug: "electronics"}}, nil
}

func (m *mockCategoryRepo) ListServiceCategories(ctx context.Context) ([]domain.ServiceCategory, error) {
	if m.onListService != nil {
		return m.onListService(ctx)
	}
	return []domain.ServiceCategory{{ID: 1, Name: "Design", Slug: "design"}}, nil
}

func (m *mockCategoryRepo) ListEventCategories(ctx context.Context) ([]domain.EventCategory, error) {
	if m.onListEvent != nil {
		return m.onListEvent(ctx)
	}
	return []domain.EventCategory{{ID: 1, Name: "Wedding", Slug: "wedding"}}, nil
}

func (m *mockCategoryRepo) ListVendorSubCategories(ctx context.Context) ([]domain.VendorSubCategory, error) {
	if m.onListVendor != nil {
		return m.onListVendor(ctx)
	}
	return []domain.VendorSubCategory{{ID: 1, CategoryID: 1, Name: "CATERING", Slug: "catering"}}, nil
}

// ---------------------------------------------------------------------------
// MidtransClient
// ---------------------------------------------------------------------------

// mockMidtransClient default: CreateSnapTransaction returns a dummy token,
// VerifySignature always returns true. Override the field for specific cases.
type mockMidtransClient struct {
	domain.MidtransClient
	onCreateSnap func(ctx context.Context, req domain.SnapRequest) (domain.SnapResult, error)
	onVerify     func(orderID, statusCode, grossAmount, signatureKey string) bool
}

func (m *mockMidtransClient) CreateSnapTransaction(ctx context.Context, req domain.SnapRequest) (domain.SnapResult, error) {
	if m.onCreateSnap != nil {
		return m.onCreateSnap(ctx, req)
	}
	return domain.SnapResult{Token: "snap-token-mock", RedirectURL: "https://snap.example.com/mock"}, nil
}

func (m *mockMidtransClient) VerifySignature(orderID, statusCode, grossAmount, signatureKey string) bool {
	if m.onVerify != nil {
		return m.onVerify(orderID, statusCode, grossAmount, signatureKey)
	}
	return true
}

// ---------------------------------------------------------------------------
// UnitOfWork
// ---------------------------------------------------------------------------

// mockUnitOfWork runs fn with the stores configured by the test.
// stores.Transactions/Wallets/Finance/Users can be filled with the same mock
// instances passed to the usecase constructor, so function-field overrides
// (e.g. onGetTransactionByID) still apply inside the transaction block.
type mockUnitOfWork struct {
	domain.UnitOfWork
	stores domain.TxStores
	onDo   func(ctx context.Context, fn func(ctx context.Context, stores domain.TxStores) error) error
}

func (m *mockUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, stores domain.TxStores) error) error {
	if m.onDo != nil {
		return m.onDo(ctx, fn)
	}
	return fn(ctx, m.stores)
}

// newMockUnitOfWork wraps a set of single mock repos into TxStores.
// All store fields point to the same instance -> suitable for usecases
// that use transaction/wallet/finance/user mock repos at once.
func newMockUnitOfWork(txRepo domain.TransactionRepository, walletRepo domain.WalletRepository, financeRepo domain.FinanceRepository, userRepo domain.UserRepository) *mockUnitOfWork {
	return &mockUnitOfWork{
		stores: domain.TxStores{
			Users:        userRepo,
			Wallets:      walletRepo,
			Transactions: txRepo,
			Finance:      financeRepo,
		},
	}
}

// newMockUnitOfWorkWithDisputes is newMockUnitOfWork plus a Disputes store, for
// the dispute resolution flow which locks the dispute row alongside the money move.
func newMockUnitOfWorkWithDisputes(txRepo domain.TransactionRepository, walletRepo domain.WalletRepository, financeRepo domain.FinanceRepository, disputeRepo domain.DisputeRepository) *mockUnitOfWork {
	muow := newMockUnitOfWork(txRepo, walletRepo, financeRepo, nil)
	muow.stores.Disputes = disputeRepo
	return muow
}

// ---------------------------------------------------------------------------
// FraudClient
// ---------------------------------------------------------------------------

// mockFraudClient default: safe transaction (isSafe=true, low score).
// Override onAnalyze to simulate a suspicious transaction.
type mockFraudClient struct {
	domain.FraudClient
	onAnalyze func(ctx context.Context, userID string, amount int64) (float64, bool, error)
}

func (m *mockFraudClient) AnalyzeTransactionRisk(ctx context.Context, userID string, amount int64) (float64, bool, error) {
	if m.onAnalyze != nil {
		return m.onAnalyze(ctx, userID, amount)
	}
	return 0.1, true, nil
}

// ---------------------------------------------------------------------------
// KYCClient
// ---------------------------------------------------------------------------

// mockKYCClient default: returns a high-confidence reference with no error.
// Override onVerify to simulate the AI service being unavailable.
type mockKYCClient struct {
	domain.KYCClient
	onVerify func(ctx context.Context, userID string, idCardURL string, selfieURL string, targetRole domain.UserRole) (float64, string, error)
}

func (m *mockKYCClient) VerifyIdentity(ctx context.Context, userID string, idCardURL string, selfieURL string, targetRole domain.UserRole) (float64, string, error) {
	if m.onVerify != nil {
		return m.onVerify(ctx, userID, idCardURL, selfieURL, targetRole)
	}
	return 0.92, "documents-genuine-same-person", nil
}

// ---------------------------------------------------------------------------
// Relayer
// ---------------------------------------------------------------------------

// mockRelayer default: returns a dummy tx hash without error.
type mockRelayer struct {
	domain.Relayer
	onLog  func(ctx context.Context, txID string, amount int64, buyer, seller string) (string, error)
	logged bool
	hash   string
}

func (m *mockRelayer) LogTransactionOnChain(ctx context.Context, txID string, amount int64, buyer, seller string) (string, error) {
	if m.onLog != nil {
		return m.onLog(ctx, txID, amount, buyer, seller)
	}
	m.logged = true
	m.hash = "0xmockrelayedhash"
	return m.hash, nil
}
