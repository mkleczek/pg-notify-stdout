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
	// Make sure only messages are printed.
	// Timestamps and other metadata
	// can be added by log processing tool outside.
	log.SetFlags(0)
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
		log.Println("Connecting to PostgreSQL")
		conn, err = pgx.Connect(context.Background(), "")

		if err != nil {
			if retryAfter > retryMax {
				log.Panic(err)
			}
			log.Println(err)
			time.Sleep(retryAfter)
			retryAfter = retryAfter + retryAfter
			continue
		}
		log.Println("Connected")
		retryAfter = retryMin

		_, err = conn.Exec(context.Background(), "LISTEN "+os.Args[1])
		if err != nil {
			log.Panic(err)
		}
		log.Println("Waiting for notifications")
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
