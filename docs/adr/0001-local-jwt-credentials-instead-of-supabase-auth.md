# Local JWT credentials instead of Supabase Auth

Supabase config keys are present in the repo but identity is zero-implemented — the only auth path was a dev `/users/token-test` backdoor that minted a JWT for any arbitrary user+role. We implement local email+password auth (bcrypt) inside core-service and mint our own HS256 JWT, reusing the existing `auth_middleware.go`. Supabase keys stay in config as a documented future swap.

## Considered

- **Supabase Auth** — deferred. It introduces a new external trust boundary and there is no existing implementation to extend; owning password hashing ourselves is the lower-risk path for an in-scope MVP.

## Consequences

- We own password hashing, login, and (later) reset/verification flows.
- `UserProfile` gains `Email` (unique) and `PasswordHash` (`json:"-"`) columns; `RegisterProfile` no longer trusts a client-supplied UUID.
- Swapping to Supabase Auth later would replace the `Login` usecase and the credential columns, but leave JWT verification and the rest of the system untouched.
