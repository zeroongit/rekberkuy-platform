package domain

// EventAuditResult menampung data hasil pemecahan dana hasil audit event
type EventAuditResult struct {
	PlatformFee     int64 `json:"platform_fee"`
	AmountToVendor  int64 `json:"amount_to_vendor"`
	BonusToEO       int64 `json:"bonus_to_eo"`
	RefundToPeserta int64 `json:"refund_to_peserta"`
}