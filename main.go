package main

import (
	"net"
	"rate_limiter/token"
	"sync"

	"github.com/hdt3213/godis/lib/logger"
)

type Connection struct {
	conn       net.Conn
	canProceed bool
}

func MakeConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
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
			tokenBasedRateLimiter := token.NewTokenRateLimiter(redisAddr, redisPassword)
			canProceed, err := tokenBasedRateLimiter.HandleRequest(userID)
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
