ALTER TABLE withdrawals
    DROP COLUMN IF EXISTS midtrans_cost,
    DROP COLUMN IF EXISTS midtrans_cost_actual;
