package token

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hdt3213/godis/lib/logger"
)

type TokenBasedRateLimiter struct {
	redisClient *redis.Client
	mu          sync.Mutex
}

func NewTokenRateLimiter(redisArr, redisPassword string) *TokenBasedRateLimiter {
	return &TokenBasedRateLimiter{
		redisClient: redis.NewClient(&redis.Options{
			Addr:     redisArr,
			Password: redisPassword,
			DB:       0,
		}),
	}
}

func (trl *TokenBasedRateLimiter) HandleRequest(userID string) (bool, error) {

	// unlock the mutex lock
	defer trl.mu.Unlock()

	ctx := context.Background()
	tokenKey := userID

	trl.mu.Lock()
	tokenExists, err := trl.redisClient.Exists(ctx, tokenKey).Result()

	fmt.Println("token exists: ", tokenExists)
	if err != nil {
		return false, err
	}

	if tokenExists <= 0 {
		newToken := userID
		if err := trl.redisClient.Set(ctx, newToken, 1, 1*time.Minute).Err(); err != nil {
			return false, err
		}

		return true, nil
	}

	counter, err := trl.redisClient.Incr(ctx, tokenKey).Result()

	if err != nil {
		return false, err
	}

	if counter > 2 {
		logger.Error("Too many requests...")
		return false, nil
	}

	return true, nil
}
