package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"phase3-api-architecture/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type ProductRepository struct {
	DB    *sql.DB
	Redis *redis.Client
}

var ctx = context.Background()

func (r *ProductRepository) GetAll(filter models.ProductFilter) ([]models.Product, error) {
	// Check cache
	// Contoh Key: products:page:1:limit:10:search:phone
	cacheKey := fmt.Sprintf("products:page:%d:limit:%d:search:%s", filter.Page, filter.Limit, filter.Search)
	cachedData, err := r.Redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []models.Product
		json.Unmarshal([]byte(cachedData), &products)
		return products, nil
	}

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
	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan hasil query
	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	// Simpan ke cache (5 menit)
	dataJson, _ := json.Marshal(products)
	r.Redis.Set(ctx, cacheKey, dataJson, 5*time.Minute)

	return products, nil
}

func (r *ProductRepository) GetByID(id int) (models.Product, error) {
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
	err = r.DB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if err != nil {
		return p, err
	}

	// Simpan ke cache (10 menit)
	dataJson, _ := json.Marshal(p)
	r.Redis.Set(ctx, cacheKey, dataJson, 10*time.Minute)

	return p, nil
}

func (r *ProductRepository) Create(p models.Product) error {
	query := "INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id"

	// Scan ID biar struk p punya ID setelah dibuat
	var newId int
	err := r.DB.QueryRow(query, p.Name, p.Price, p.Stock).Scan(&newId)
	if err != nil {
		return err
	}

	// Invalidate cache, karena data sudah berubah / baru
	r.Redis.Del(ctx, "products:all")

	return nil
}

func (r *ProductRepository) Update(p models.Product) error {
	query := "UPDATE products SET name = $1, price = $2, stock = $3 WHERE id = $4"
	res, err := r.DB.Exec(query, p.Name, p.Price, p.Stock, p.ID)
	if err != nil {
		return err
	}

	// Cek apakah ada baris yang terupdate
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows // Error jika tidak ada baris yang diupdate atau ID tidak ditemukan
	}

	// Invalidate cache
	r.Redis.Del(ctx, "products:all")
	r.Redis.Del(ctx, fmt.Sprintf("product:%d", p.ID))

	return nil
}

func (r *ProductRepository) Delete(id int) error {
	query := "DELETE FROM products WHERE id = $1"
	res, err := r.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	r.Redis.Del(ctx, "products:all")
	r.Redis.Del(ctx, fmt.Sprintf("product:%d", id))

	return nil
}

func (r *ProductRepository) Checkout(ctx context.Context, userID int, req models.CheckoutRequest) error {
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

	r.Redis.Del(ctx, "products:all")
	r.Redis.Del(ctx, fmt.Sprintf("product:%d", req.ProductID))

	return nil
}
