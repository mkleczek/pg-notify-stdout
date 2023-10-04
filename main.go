package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
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
	retryCount := 1
	var conn *pgx.Conn
	var err error
	defer func() {
		if conn != nil {
			conn.Close(context.Background())
		}
	}()

	brokenConnectionOutput := os.NewFile(3, "")
	for {
		log.Printf("Trying to connect to database. Attempt %v\n", retryCount)
		conn, err = pgx.Connect(context.Background(), "")

		if err != nil {
			log.Println(err)
			log.Printf("Next database connection attempt in %v\n", retryAfter)
			time.Sleep(retryAfter)
			// Exponential backoff with equal jitter
			retryAfter = (1 << retryCount) * retryMin
			if retryAfter > retryMax {
				retryAfter = retryMax
			}
			retryMilis := retryAfter.Milliseconds() / 2
			retryMilis = retryMilis + (rand.Int63n(1000) * retryMilis / 1000)
			retryAfter = time.Duration(retryMilis) * time.Millisecond
			continue
		}
		log.Println("Connected to database")
		retryAfter = retryMin
		retryCount = 1

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
				brokenConnectionOutput.WriteString("\n")
				break
			}

			fmt.Println(notification.Payload)
		}
	}
}
