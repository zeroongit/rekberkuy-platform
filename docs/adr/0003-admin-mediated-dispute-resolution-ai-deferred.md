# Admin-mediated dispute resolution; AI auto-resolve deferred

The product docs describe "AI resolution" of disputes, but the AI service (`backend-ai`, out of scope for this build) does not exist in the repo. We resolve disputes by an Admin who sets a binding `Outcome` of `REFUND_BUYER` or `RELEASE_TO_SELLER`. No automated AI resolution ships this session.

## Considered

- **AI-assisted auto-resolution** — deferred until `backend-ai` exists. Shipping it now would call a service that isn't there.

## Consequences

- Disputes gain an explicit `DisputeStatus` (`OPEN → UNDER_REVIEW → RESOLVED`) and an `Outcome`; `IsResolved` becomes a derived convenience.
- The resolution usecase moves money within the `UnitOfWork`: `REFUND_BUYER` credits the buyer's wallet and ends the transaction `REFUNDED`; `RELEASE_TO_SELLER` credits the seller and ends it `RELEASED`.
- Any future AI advisory layer sits *on top of* the Admin decision, not in place of it.
