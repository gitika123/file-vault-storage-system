package ratelimit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter provides a shared fixed-window limit across API replicas.
type RedisLimiter struct {
	client *redis.Client
	rate   int
	burst  int
}

func NewRedis(client *redis.Client, perSecond, burst int) *RedisLimiter {
	return &RedisLimiter{client: client, rate: perSecond, burst: burst}
}

func (l *RedisLimiter) Allow(key string) bool {
	if l == nil || l.client == nil {
		return true
	}
	now := time.Now().Unix()
	window := now
	hash := sha256.Sum256([]byte(key))
	redisKey := "balkanid:ratelimit:" + hex.EncodeToString(hash[:]) + ":" + time.Unix(now, 0).Format("20060102150405")
	limit := l.rate
	if l.burst > limit {
		limit = l.burst
	}
	count, err := l.client.Incr(context.Background(), redisKey).Result()
	if err != nil {
		return true
	}
	if count == 1 {
		_ = l.client.ExpireAt(context.Background(), redisKey, time.Unix(window+1, 0)).Err()
	}
	return count <= int64(limit)
}
