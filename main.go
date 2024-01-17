package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hdt3213/godis/lib/logger"
)

type TokenBasedRateLimiter struct {
	redisClient *redis.Client
	mu          sync.Mutex
}

type Connection struct {
	conn       net.Conn
	canProceed bool
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

func MakeConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

func (trl *TokenBasedRateLimiter) handleRequest(userID string) (bool, error) {

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

func main() {
	var wg sync.WaitGroup

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		logger.Error(err)
	}

	ch := make(chan *Connection)
	defer close(ch)

	for {

		conn, err := listener.Accept()

		if err != nil {
			logger.Error(err)
		}
		c := MakeConnection(conn)

		go func(conn net.Conn) {
			wg.Add(1)

			defer wg.Done()
			redisAddr := "localhost:49153"
			redisPassword := "redispw"
			userID := "user123"
			tokenBasedRateLimiter := NewTokenRateLimiter(redisAddr, redisPassword)
			canProceed, err := tokenBasedRateLimiter.handleRequest(userID)
			if err != nil {
				logger.Error(err)
			}

			c.canProceed = canProceed
			ch <- c

		}(conn)

		go func() {
			for c := range ch {
				if c.canProceed {
					// pass the request to the backend service
					logger.Info("Request Allowed.")

				} else {
					c.conn.Write([]byte("Request Not Allowed"))
					c.conn.Close()
					logger.Error("Request is not Allowed")
				}
			}
		}()

	}

}
