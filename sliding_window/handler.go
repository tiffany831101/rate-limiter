package slidingWindow

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hdt3213/godis/lib/logger"
)

type SlidingWindow struct {
	redisClient *redis.Client
	mu          sync.Mutex
	windowSize  int
}

func NewSlidingWindow(redisArr string, redisPassword string, windowSize int) *SlidingWindow {
	return &SlidingWindow{
		redisClient: redis.NewClient(&redis.Options{Addr: redisArr, Password: redisPassword, DB: 0}),
		windowSize:  windowSize,
	}
}

func (sw *SlidingWindow) HandleRequest(ctx context.Context, key string) (bool, error) {

	currentTime := time.Now()
	op := &redis.ZRangeBy{
		Min:    strconv.Itoa(int(currentTime.Add(-2 * time.Second).Unix())),
		Max:    strconv.Itoa(int(currentTime.Unix())),
		Offset: 0,
	}

	rangeRequests, err := sw.redisClient.ZRevRangeByScore(ctx, key, op).Result()

	if err != nil {
		logger.Error(err)
		return false, err
	}

	fmt.Println(len(rangeRequests))
	if len(rangeRequests) >= sw.windowSize {
		return false, nil
	}

	score := float64(time.Now().Unix())
	_, e := sw.redisClient.ZAdd(ctx, key, &redis.Z{Score: score, Member: score}).Result()
	if e != nil {
		logger.Error(err)
	}

	return true, nil
}
