package bucket

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hdt3213/godis/lib/logger"
)

// SharedBucketRateLimiter represents a rate limiter based on a shared bucket in a sorted set.
type SharedBucketRateLimiter struct {
	redisClient   *redis.Client
	currentTokens int32
	maxToken      int32
	mu            sync.Mutex
}

// NewSharedBucketRateLimiter creates a new instance of SharedBucketRateLimiter.
func NewSharedBucketRateLimiter(redisArr, redisPassword string, currentToken, maxToken int32) *SharedBucketRateLimiter {
	return &SharedBucketRateLimiter{
		redisClient:   redis.NewClient(&redis.Options{Addr: redisArr, Password: redisPassword, DB: 0}),
		currentTokens: currentToken,
		maxToken:      maxToken,
	}
}

// AddTokenBackgroundProcess runs in the background to add tokens to the shared bucket.
func (sbrl *SharedBucketRateLimiter) AddTokenBackgroundProcess(key string) {
	ctx := context.Background()

	for {
		sbrl.mu.Lock()
		fmt.Println(sbrl.currentTokens, sbrl.maxToken)

		if sbrl.currentTokens < sbrl.maxToken {

			// Add tokens to the sorted set
			score := float64(time.Now().Unix() + 1)
			atomic.AddInt32((*int32)(&sbrl.currentTokens), 1)

			// Add the token to the Sorted Set
			_, err := sbrl.redisClient.ZAdd(ctx, key, &redis.Z{Score: score, Member: score}).Result()
			if err != nil {
				logger.Error(err)
			}
		}
		sbrl.mu.Unlock()

		time.Sleep(time.Second)
	}
}

// HandleRequest processes incoming requests based on token availability.
func (sbrl *SharedBucketRateLimiter) HandleRequest(ctx context.Context, key string) (bool, error) {
	sbrl.mu.Lock()
	defer sbrl.mu.Unlock()

	// Get the token from the sorted set
	result, err := sbrl.redisClient.ZRangeWithScores(ctx, key, 0, -1).Result()
	if err != nil {
		logger.Error(err)
		return false, err
	}

	if len(result) > 0 {
		member := result[0].Member
		score := result[0].Score

		if float64(time.Now().Unix()) < score {
			logger.Error("Request is not available now")
			// Return or handle the unavailability
			return false, nil
		} else {
			// Remove the token from the sorted set
			_, err := sbrl.redisClient.ZRem(ctx, key, member).Result()
			if err != nil {
				logger.Error(err)
				return false, err
			}

			// Update current tokens
			sbrl.currentTokens--
			logger.Info("Request is available")

			return true, nil
		}
	}

	return false, nil
}
