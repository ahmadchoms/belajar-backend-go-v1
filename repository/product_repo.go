package repository

import (
	"context"
	"database/sql"
	"encoding/json"
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

func (r *ProductRepository) GetAll() ([]models.Product, error) {
	rows, err := r.DB.Query("SELECT  * FROM products")

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Stock); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	time.Sleep(5 * time.Second)
	return products, nil
}

func (r *ProductRepository) Create(p models.Product) (models.Product, error) {
	query := "INSERT INTO products (name, price, stock) VALUES ($1, $2, $3) RETURNING id"
	err := r.DB.QueryRow(query, p.Name, p.Price, p.Stock).Scan(&p.ID)
	return p, err
}

func (r *ProductRepository) GetByID(id int) (models.Product, error) {
	var p models.Product

	key := fmt.Sprintf("product:%d", id)

	val, err := r.Redis.Get(ctx, key).Result()

	if err == nil {
		errUnmarshal := json.Unmarshal([]byte(val), &p)
		if errUnmarshal == nil {
			fmt.Println("🚀 Kena Cache Redis! Gak perlu ke DB.")
			return p, nil
		}
	} else if err != redis.Nil {
		fmt.Println("⚠️ Redis Error:", err)
	}

	fmt.Println("🐢 Cache Miss. Ambil dari DB...")
	query := "SELECT id, name, price, stock FROM products WHERE id = $1"
	err = r.DB.QueryRow(query, id).Scan(&p.ID, &p.Name, &p.Price, &p.Stock)
	if err != nil {
		return p, err
	}

	jsonBytes, _ := json.Marshal(p)

	err = r.Redis.Set(ctx, key, jsonBytes, 1*time.Minute).Err()
	if err != nil {
		fmt.Println("Gagal simpan ke Redis:", err)
	}

	return p, nil
}
