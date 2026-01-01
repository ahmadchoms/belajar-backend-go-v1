package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"phase3-api-architecture/internal/event"
	"phase3-api-architecture/internal/worker"
	"phase3-api-architecture/models"
	"phase3-api-architecture/pkg/resiliency"
	"phase3-api-architecture/pkg/stream"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker"
)

type ProductRepository struct {
	DB    *sql.DB
	Redis *redis.Client
	// Tambahan v4
	Breaker *gobreaker.CircuitBreaker
	Kafka   *stream.KafkaProducer
}

// tambahkan contructor (agar breaker ter-inisialisasi)
func NewProductRepository(db *sql.DB, rdb *redis.Client, kafka *stream.KafkaProducer) *ProductRepository {
	return &ProductRepository{
		DB:      db,
		Redis:   rdb,
		Breaker: resiliency.NewDatabaseBreaker("product-db-query"),
		Kafka:   kafka,
	}
}

func (r *ProductRepository) GetAll(ctx context.Context, filter models.ProductFilter) ([]models.Product, error) {
	// Check cache
	// Contoh Key: products:page:1:limit:10:search:phone
	cacheKey := fmt.Sprintf("products:page:%d:limit:%d:search:%s", filter.Page, filter.Limit, filter.Search)
	cachedData, err := r.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []models.Product
		json.Unmarshal([]byte(cachedData), &products)
		return products, nil
	}

	result, err := r.Breaker.Execute(func() (interface{}, error) {
		// Build query dengan filter
		query := "SELECT id, name, price, stock FROM products WHERE 1=1"
		var args []interface{}
		argCounter := 1

		// Tambahkan filter pencarian jika ada
		if filter.Search != "" {
			query += fmt.Sprintf(" AND name ILIKE $%d", argCounter)
			args = append(args, "%"+filter.Search+"%")
			argCounter++
		}

		// Tambahkan pagination
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCounter, argCounter+1)
		args = append(args, filter.Limit, filter.GetOffset())

		// Ambil data dari database
		rows, err := r.DB.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		// Scan hasil query
		var products []models.Product
		for rows.Next() {
			var p models.Product
			if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
				return nil, err
			}
			products = append(products, p)
		}

		return products, nil
	})

	// handle error dari breaker
	if err != nil {
		fmt.Printf("[DEBUG REPO] Error from breaker: %v | Type: %T\n", err, err)
		if err == gobreaker.ErrOpenState {
			// sirkuit putus! jangan pakai DB
			fmt.Println("[DEBUG REPO] CIRCUIT IS OPEN! Returning ErrServiceUnavailable")
			return nil, resiliency.ErrServiceUnavailbale
		}
		return nil, err
	}

	// casting result interface{} kembali ke tipe asli
	products := result.([]models.Product)

	// Simpan ke cache (5 menit)
	dataJson, _ := json.Marshal(products)
	r.Redis.Set(ctx, cacheKey, dataJson, 5*time.Minute)

	return products, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (models.Product, error) {
	var p models.Product
	cacheKey := fmt.Sprintf("product:%d", id)

	// Check cache data berdasarkan ID
	cachedData, err := r.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		json.Unmarshal([]byte(cachedData), &p)
		return p, nil
	}

	// Ambil data dari database
	query := "SELECT id, name, price, stock FROM products WHERE id = $1"
	err = r.DB.QueryRowContext(ctx, query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if err != nil {
		return p, err
	}

	// Simpan ke cache (10 menit)
	dataJson, _ := json.Marshal(p)
	r.Redis.Set(ctx, cacheKey, dataJson, 10*time.Minute)

	return p, nil
}

func (r *ProductRepository) Create(ctx context.Context, p *models.Product) error {
	// simpan ke db
	query := "INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id"
	err := r.DB.QueryRowContext(ctx, query, p.Name, p.Price, p.Stock).Scan(&p.ID)
	if err != nil {
		return err
	}

	// hapus cache
	r.Redis.Del(ctx, "products:all")

	// kirim event ke kafka
	// kirim ke topic "product-events" agar worker elasticsearch menangkapnya
	evt := event.ProductEvent{
		Action:  event.ActionCreate,
		Product: *p,
	}

	// Gunakan ID sebagai Key agar urutan event untuk produk ini terjamin
	if err := r.Kafka.SendMessage("product-events", fmt.Sprintf("%d", p.ID), evt); err != nil {
		fmt.Printf("[WARNING] Gagal kirim event create ke Kafka: %v\n", err)
	}

	return nil
}

func (r *ProductRepository) Update(ctx context.Context, p *models.Product) error {
	// 1. Update DB (Code Lama)
	query := "UPDATE products SET name=$1, price=$2, stock=$3 WHERE id=$4"
	_, err := r.DB.ExecContext(ctx, query, p.Name, p.Price, p.Stock, p.ID)
	if err != nil {
		return err
	}

	// 2. Hapus Cache (Code Lama)
	r.Redis.Del(ctx, "products:all")
	r.Redis.Del(ctx, fmt.Sprintf("product:%d", p.ID))

	// 3. [BARU] KIRIM EVENT KE KAFKA
	evt := event.ProductEvent{
		Action:  event.ActionUpdate,
		Product: *p,
	}
	if err := r.Kafka.SendMessage("product-events", fmt.Sprintf("%d", p.ID), evt); err != nil {
		fmt.Printf("[WARNING] Gagal kirim event update ke Kafka: %v\n", err)
	}

	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	// 1. Delete DB (Code Lama)
	query := "DELETE FROM products WHERE id = $1"
	_, err := r.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// 2. Hapus Cache (Code Lama)
	r.Redis.Del(ctx, "products:all")
	r.Redis.Del(ctx, fmt.Sprintf("product:%d", id))

	// 3. [BARU] KIRIM EVENT KE KAFKA
	// Payload produk kosong, cukup ID-nya saja yang penting untuk event delete
	evt := event.ProductEvent{
		Action:  event.ActionDelete,
		Product: models.Product{ID: id},
	}
	if err := r.Kafka.SendMessage("product-events", fmt.Sprintf("%d", id), evt); err != nil {
		fmt.Printf("[WARNING] Gagal kirim event delete ke Kafka: %v\n", err)
	}

	return nil
}

func (r *ProductRepository) Checkout(ctx context.Context, userID int, userEmail string, req models.CheckoutRequest) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	queryUpdate := `
		UPDATE products 
		SET stock = stock - $1 
		WHERE id = $2 AND stock >= $1 
		RETURNING price`

	var pricePerItem int
	err = tx.QueryRowContext(ctx, queryUpdate, req.Quantity, req.ProductID).Scan(&pricePerItem)

	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("stok tidak mencukupi atau produk tidak ditemukan")
		}
		return err
	}

	totalPrice := pricePerItem * req.Quantity

	queryInsert := `
		INSERT INTO transactions (user_id, product_id, quantity, total_price) 
		VALUES ($1, $2, $3, $4)`

	_, err = tx.ExecContext(ctx, queryInsert, userID, req.ProductID, req.Quantity, totalPrice)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	task := worker.TaskSendInvoice{
		UserID:     userID,
		Email:      userEmail,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: totalPrice,
	}

	err = r.Kafka.SendMessage("checkout-events", fmt.Sprintf("%d", userID), task)
	if err != nil {
		// Jika Kafka mati, transaction DB sudah terlanjur commit.
		// Di sistem distributed yang ketat, butuh "Outbox Pattern".
		log.Printf("[ERROR] Gagal kirim event ke Kafka: %v", err)
		// Tidak return error agar user tetap tau checkout berhasil (meski email telat)
	}

	return nil
}
