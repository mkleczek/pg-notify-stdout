package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func main() {
	const retryMin = 1 * time.Second
	const retryMax = 1 * time.Hour
	retryAfter := retryMin
	var conn *pgx.Conn
	var err error
	defer func() {
		if conn != nil {
			conn.Close(context.Background())
		}
	}()
	for {
		conn, err = pgx.Connect(context.Background(), "")

		if err != nil {
			time.Sleep(retryAfter)
			retryAfter = retryAfter + retryAfter
			if retryAfter > retryMax {
				log.Panic(err)
			}
			continue
		}
		retryAfter = retryMin

		_, err = conn.Exec(context.Background(), "LISTEN "+os.Args[1])
		if err != nil {
			log.Panic(err)
		}
		for {
			notification, err := conn.WaitForNotification(context.Background())
			if err != nil {
				conn.Close(context.Background())
				conn = nil
				log.Println(err)
				break
			}

			fmt.Println(notification.Payload)
		}
	}
}
