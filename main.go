package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"phase3-api-architecture/handler"
	"phase3-api-architecture/repository"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "root"
	}

	dbPass := os.Getenv("DB_PASSWORD")
	if dbPass == "" {
		dbPass = "rahasia"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "belajargo"
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	query := `
	CREATE TABLE IF NOT EXISTS products (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		price INT NOT NULL,
		stock INT NOT NULL
	)`
	_, err = db.Exec(query)
	if err != nil {
		log.Fatal("Gagal membuat tabel", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	productRepo := &repository.ProductRepository{DB: db, Redis: rdb}
	productHandler := &handler.ProductHandler{Repo: productRepo}

	mux := http.NewServeMux()
	mux.HandleFunc("/products", productHandler.HandlerProducts)
	mux.HandleFunc("/products/", productHandler.HandleProductByID)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		fmt.Println("🚀 Server Phase 7 running at :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server Crash: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	fmt.Println("\n⚠️  Server sedang dimatikan...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Mwahahahahah")
	}

	if err := db.Close(); err != nil {
		fmt.Println("Gagal tutup DB:", err)
	}
	if err := rdb.Close(); err != nil {
		fmt.Println("Gagal tutup Redis:", err)
	}

	fmt.Println("✅ Server mati dengan tenang (Graceful Shutdown).")
}
