package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"phase3-api-architecture/internal/worker"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	log.Printf("Worker Invoice Starting...")

	// Setup Redis client
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	// Cek koneksi ke Redis
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("Gagal konek ke Redis")
	}

	// Handle graceful shutdown, biar gak ada goroutine yang nyangkut
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// Loop Abadi
	go func() {
		for {
			if err := worker.ProcessTaskInvoice(ctx, rdb); err != nil {
				log.Printf("Critical Error: %v", err)
				time.Sleep(1 * time.Second)
			}
		}
	}()

	// Tunggu sinyal untuk shutdown
	<-quit
	log.Println("Worker shutting down...")
	rdb.Close()
	log.Println("Worker Invoice Stopped")

}
