package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

func ProcessTaskInvoice(ctx context.Context, rdb *redis.Client) error {
	result, err := rdb.BLPop(ctx, 0, QueueInvoice).Result()
	if err != nil {
		return fmt.Errorf("redis connection error: %v", err)
	}

	payload := result[1]

	var task TaskSendInvoice
	if err := json.Unmarshal([]byte(payload), &task); err != nil {
		log.Printf("Data rusak: %s", payload)
		return nil
	}

	log.Printf("[START] Mengirim Invoice ke %s", task.Email)

	if rand.Intn(10) < 2 {
		log.Printf("[Error] Gagal connect ke Emaill Server untuk %s", task.Email)

		if err := rdb.RPush(ctx, QueueInvoice, payload).Err(); err != nil {
			return fmt.Errorf("Gagal mengembalikan task ke redis: %v", err)
		}

		log.Printf("[RETRY] Task dikembalikan ke antrian")
		return nil
	}

	time.Sleep(2 * time.Second)

	log.Printf("[DONE] Berhasil mengirim Invoice ke %d\n", task.UserID)
	return nil

}
