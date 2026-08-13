# Admin-triggered disbursement instead of the Midtrans Disbursement API

External event vendors are modelled as `PENDING_DISBURSEMENT`, but actually paying them requires Midtrans Disbursement — a separate gateway product with its own enrolment and API keys that we cannot enrol or verify in this environment. We model the external-vendor payout as an Admin action: an Admin reviews the payout (bank details + invoice URL already on `EventVendorPayout`), performs the transfer out-of-band, then marks the payout `DISBURSED`. The `mark-disbursed` call records completion only — it moves no money in-system.

## Considered

- **Real Midtrans Disbursement API now** — blocked: separate product, no enrolment/keys, nothing to verify against.
- **Background disbursement worker** — premature; there is no batching need and it would hide a money-out action from human review.

## Consequences

- New terminal status `DISBURSED` for external payouts (today only `PENDING`/`APPROVED`/`PENDING_DISBURSEMENT` exist).
- `IsDisbursedByMidtrans` stays `false` on the admin path; it flips to `true` only if/when the real API is wired behind this same usecase later — callers don't change.
