package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"rekberkuy/core-service/internal/domain"
)

// idempotencyRedisRepository stores idempotency keys in Redis using
// SET NX (atomic) + automatic TTL. Faster & does not burden PostgreSQL.
type idempotencyRedisRepository struct {
	client *redis.Client
	ttl    time.Duration
}

// NewIdempotencyRedisRepository takes an active redis client and a TTL in seconds.
func NewIdempotencyRedisRepository(client *redis.Client, ttlSec int) domain.IdempotencyRepository {
	return &idempotencyRedisRepository{
		client: client,
		ttl:    time.Duration(ttlSec) * time.Second,
	}
}

func key(id string) string { return "idem:" + id }

// CheckOrLock uses SET NX atomically:
//   - set succeeds   -> new request (isNew=true)
//   - key exists     -> duplicate transaction, fetch old payload (isNew=false)
func (r *idempotencyRedisRepository) CheckOrLock(ctx context.Context, rec *domain.IdempotencyRecord) (*domain.IdempotencyRecord, bool, error) {
	payload, err := json.Marshal(rec)
	if err != nil {
		return nil, false, fmt.Errorf("failed to marshal idempotency record: %w", err)
	}

	ok, err := r.client.SetNX(ctx, key(rec.ID), payload, r.ttl).Result()
	if err != nil {
		return nil, false, fmt.Errorf("failed to setnx redis idempotency: %w", err)
	}

	// New request successfully locked
	if ok {
		return rec, true, nil
	}

	// Key already exists -> fetch stored record
	raw, err := r.client.Get(ctx, key(rec.ID)).Bytes()
	if err != nil {
		return nil, false, fmt.Errorf("failed to read existing idempotency record: %w", err)
	}

	var existing domain.IdempotencyRecord
	if err := json.Unmarshal(raw, &existing); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal existing idempotency record: %w", err)
	}
	return &existing, false, nil
}

// SaveResponse overwrites the idempotency record with the actual response from the first request.
// The TTL is refreshed so the dedup window stays consistent from the last execution.
func (r *idempotencyRedisRepository) SaveResponse(ctx context.Context, id string, status int, body []byte) error {
	raw, err := r.client.Get(ctx, key(id)).Bytes()
	if err != nil {
		return fmt.Errorf("failed to read idempotency record for update: %w", err)
	}
	var rec domain.IdempotencyRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		return fmt.Errorf("failed to unmarshal idempotency record for update: %w", err)
	}
	rec.ResponseBody = body
	rec.ResponseStatus = status
	payload, err := json.Marshal(rec)
	if err != nil {
		return fmt.Errorf("failed to marshal idempotency record: %w", err)
	}
	if err := r.client.Set(ctx, key(id), payload, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to save idempotency response: %w", err)
	}
	return nil
}
