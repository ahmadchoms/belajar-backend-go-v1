package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"phase3-api-architecture/handler"
	"phase3-api-architecture/middleware"
	pb "phase3-api-architecture/pb/proto/inventory"
	"phase3-api-architecture/pkg/stream"
	"phase3-api-architecture/pkg/telemetry"
	"phase3-api-architecture/repository"
	"strings"
	"syscall"
	"time"

	_ "phase3-api-architecture/docs"

	"github.com/XSAM/otelsql"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// Init Tracing
	// Hubungkan ke OTel Collector
	collectorAddr := os.Getenv("OTEL_COLLECTOR_ADDR")
	shutdown := telemetry.InitTracer("inventory-api", collectorAddr)
	defer func() {
		if err := shutdown(context.Background()); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()

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

	// Database Instrumentation
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := otelsql.Open("postgres", connStr, otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL))
	if err != nil {
		log.Fatal(err)
	}

	// Register Metrics DB (Connection Pool stats otomatis dikirim ke Prometheus)
	shutdownDBMetrics, err := otelsql.RegisterDBStatsMetrics(
		db,
		otelsql.WithAttributes(semconv.DBSystemNamePostgreSQL),
	)
	if err != nil {
		log.Printf("Gagal register db metrics: %v", err)
	}

	defer func() {
		if err := shutdownDBMetrics.Unregister(); err != nil {
			log.Printf("Gagal unregister DB metrics: %v", err)
		}
	}()

	defer db.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", redisHost, redisPort),
	})

	// Enable Tracing di Redis
	// if err := redisotel.InstrumentTracing(rdb); err != nil {
	// 	log.Fatal(err)
	// }

	// Enable Metrics di Redis
	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		log.Fatal(err)
	}

	kafkaAddr := os.Getenv("KAFKA_BROKERS")
	if kafkaAddr == "" {
		kafkaAddr = "kafka:9092" // Default untuk docker-compose
	}
	kafkaBrokers := strings.Split(kafkaAddr, ",")

	// Gunakan variabel kafkaBrokers yg didapat dari Env
	kafkaProducer := stream.NewKafkaProducer(kafkaBrokers)
	defer kafkaProducer.Close()

	productRepo := repository.NewProductRepository(db, rdb, kafkaProducer)
	productHandler := &handler.ProductHandler{Repo: productRepo}
	userRepo := &repository.UserRepository{DB: db}
	authHandler := &handler.AuthHandler{Repo: userRepo}

	// Allow 20 request/detik, dengan burst maksimal 30
	rateLimitter := middleware.NewIPRateLimiter(rate.Limit(20), 30)

	go func() {
		lis, err := net.Listen("tcp", "[::]:50051") // Port gRPC biasanya
		if err != nil {
			log.Fatalf("Gagal listen port 50051: %v", err)
		}

		grpcServer := grpc.NewServer(
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		)

		// Register Handler ke Server gRPC
		inventoryGrpcHandler := &handler.GrpcInventoryHandler{Repo: productRepo}
		pb.RegisterInventoryServiceServer(grpcServer, inventoryGrpcHandler)

		reflection.Register(grpcServer)

		fmt.Println("gRPC Server Listening at :50051")
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Gagal start gRPC: %v", err)
		}
	}()

	// HTTP Router
	mux := http.NewServeMux()

	// Health Check Endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

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

	// Gunakan fungsi spesifik 'HandleCheckout'
	mux.Handle("POST /checkout", stackAuth(http.HandlerFunc(productHandler.HandleCheckout)))

	// --- 3. ADMIN ROUTES ---
	// Create
	mux.Handle("POST /products", stackAdmin(http.HandlerFunc(productHandler.HandleCreateProduct)))

	// Update (PUT)
	mux.Handle("PUT /products/{id}", stackAdmin(http.HandlerFunc(productHandler.HandleUpdateProduct)))

	// Delete (DELETE)
	mux.Handle("DELETE /products/{id}", stackAdmin(http.HandlerFunc(productHandler.HandleDeleteProduct)))

	// Otomatis membuat "Span" untuk setiap req HTTP yang masuk
	otelHandler := otelhttp.NewHandler(mux, "server-root")
	finalHandler := rateLimitter.Limit(otelHandler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      finalHandler,
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

	rdb.Close()
	fmt.Println("✅ Server mati dengan tenang.")
}
