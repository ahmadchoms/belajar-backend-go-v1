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
	"phase3-api-architecture/middleware"
	"phase3-api-architecture/repository"
	"syscall"
	"time"

	_ "phase3-api-architecture/docs"

	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title           Inventory API
// @version         2.0
// @description     API untuk manajemen inventory warung (Backend Engineering v2).
// @termsOfService  http://swagger.io/terms/

// @contact.name    Ahmadchoms
// @contact.email   ahmad@example.com

// @host            localhost:8080
// @BasePath        /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	middleware.InitLogger()

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "8080"
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

	queries := []string{`
		CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			price INT NOT NULL,
			stock INT NOT NULL
		)`,
		`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range queries {
		_, err := db.Exec(q)
		if err != nil {
			log.Fatalf("Gagal menjalankan query: %v\nError: %v", q, err)
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	productRepo := &repository.ProductRepository{DB: db, Redis: rdb}
	productHandler := &handler.ProductHandler{Repo: productRepo}
	userRepo := &repository.UserRepository{DB: db}
	authHandler := &handler.AuthHandler{Repo: userRepo}

	mux := http.NewServeMux()

	// Logger Only
	stackLogger := func(h http.Handler) http.Handler {
		return middleware.LoggerMiddleware(h)
	}
	// Logger + Auth (User Biasa Boleh Masuk)
	stackAuth := func(h http.Handler) http.Handler {
		return middleware.LoggerMiddleware(middleware.AuthMiddleware(h))
	}
	// Logger + Auth + Admin (Hanya Admin)
	stackAdmin := func(h http.Handler) http.Handler {
		return middleware.LoggerMiddleware(
			middleware.AuthMiddleware(
				middleware.AdminMiddleware(h),
			),
		)
	}

	// --- 1. PUBLIC ROUTES ---
	mux.Handle("POST /register", stackLogger(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /login", stackLogger(http.HandlerFunc(authHandler.Login)))

	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	// --- 2. USER ROUTES ---
	// Gunakan fungsi spesifik 'GetAllProducts' (bukan dispatcher HandlerProducts)
	mux.Handle("GET /products", stackAuth(http.HandlerFunc(productHandler.GetAllProducts)))

	// Gunakan fungsi spesifik 'GetProductByID'
	mux.Handle("GET /products/{id}", stackAuth(http.HandlerFunc(productHandler.HandleGetProductByID)))

	// --- 3. ADMIN ROUTES ---
	// Create
	mux.Handle("POST /products", stackAdmin(http.HandlerFunc(productHandler.HandleCreateProduct)))

	// Update (PUT)
	mux.Handle("PUT /products/{id}", stackAdmin(http.HandlerFunc(productHandler.HandleUpdateProduct)))

	// Delete (DELETE)
	mux.Handle("DELETE /products/{id}", stackAdmin(http.HandlerFunc(productHandler.HandleDeleteProduct)))

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
		log.Println("Server forced to shutdown:", err)
	}

	db.Close()
	rdb.Close()
	fmt.Println("✅ Server mati dengan tenang.")
}
