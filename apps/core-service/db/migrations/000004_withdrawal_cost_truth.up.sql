-- 000004: withdrawal disbursement cost tracking (true-up model).
-- midtrans_cost       = the ESTIMATED Midtrans disbursement cost booked at
--                       request time (from MIDTRANS_DISBURSEMENT_FEE).
-- midtrans_cost_actual = the REAL cost, recorded when the admin confirms the
--                       disbursement (or when the future DisbursementClient
--                       adapter reports it from the API response). The ledger
--                       reconciles the difference at that moment.

ALTER TABLE withdrawals
    ADD COLUMN midtrans_cost        BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN midtrans_cost_actual BIGINT;
