package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"phase3-api-architecture/internal/event"
	"phase3-api-architecture/internal/worker"
	"phase3-api-architecture/pkg/search"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/elastic/go-elasticsearch/v7"
)

func main() {
	// Konfigurasi
	brokers := os.Getenv("KAFKA_BROKERS")
	esAddress := os.Getenv("ELASTICSEARCH_ADDRESS")
	if brokers == "" {
		brokers = "kafka:9092"
	}
	if esAddress == "" {
		esAddress = "http://elasticsearch:9200" // Default docker internal
	}
	brokerList := strings.Split(brokers, ",")
	groupID := "inventory-worker-group"
	esClient := search.InitES(esAddress)

	// 2. Setup Sarama Config
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest // Baca dari awal jika belum ada offset

	// Inject ES Client ke Handler
	consumer := &ConsumerHandler{
		esClient: esClient,
	}

	// 3. Init Consumer Group
	client, err := sarama.NewConsumerGroup(brokerList, groupID, config)
	if err != nil {
		log.Panicf("[KAFKA-WORKER] Error creating consumer group client: %v", err)
	}

	// 4. Handle OS Signals (Graceful Shutdown)
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			topics := []string{"checkout-events", "product-events"}

			if err := client.Consume(ctx, topics, consumer); err == nil {
				log.Panicf("[KAFKA-WORKER] Error from consumer: %v", err)
			}
			// Cek apakah context dibatalkan (aplikasi mau mati)
			if ctx.Err() != nil {
				return
			}
		}
	}()

	log.Println("🚀 Kafka Worker Started! Listening to 'checkout-events'...")

	<-sigterm // Block disini sampai ada CTRL+C
	log.Println("⚠️  Shutdown signal received, closing worker...")

	cancel()  // Beritahu semua goroutine untuk berhenti
	wg.Wait() // Tunggu sampai cleanup selesai

	if err = client.Close(); err != nil {
		log.Panicf("[KAFKA-WORKER] Error closing client: %v", err)
	}
}

type AuditLog struct {
	Timestamp time.Time   `json:"timestamp"`
	Action    string      `json:"action"`
	ProductID int         `json:"product_id"`
	Payload   interface{} `json:"payload,omitempty"`
}

// --- CONSUMER HANDLER ---

type ConsumerHandler struct {
	esClient *elasticsearch.Client // Worker punya akses ke ES
}

func (h *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	log.Println("[KAFKA-WORKER] Partition assigned")
	return nil
}

func (h *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	log.Println("[KAFKA-WORKER] Partition revoked")
	return nil
}

func (h *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("[KAFKA-WORKER] Got message topic=%s partition=%d offset=%d", message.Topic, message.Partition, message.Offset)

		// Routing berdasarkan TOPIC
		switch message.Topic {
		case "checkout-events":
			var task worker.TaskSendInvoice
			if err := json.Unmarshal(message.Value, &task); err != nil {
				processTask(task) // Logic lama kirim email
			}

		case "product-events":
			var evt event.ProductEvent
			if err := json.Unmarshal(message.Value, &evt); err != nil {
				log.Printf("[ERROR] Gagal parse product event: %v", err)
				continue
			}

			// panggil fungsi singkronisasi ke ES
			h.syncProductToES(evt)
		}

		session.MarkMessage(message, "")
	}

	return nil
}

// Logic pemrosesan (bisa dipindah ke internal/worker/processor.go agar lebih rapi)
func processTask(t worker.TaskSendInvoice) {
	// Simulasi kerja berat
	log.Printf("📧 Sending Invoice to %s for ProductID %d (Total: %d)...", t.Email, t.ProductID, t.TotalPrice)
	time.Sleep(1 * time.Second)
	log.Println("✅ Invoice Sent Successfully!")
}

func (h *ConsumerHandler) syncProductToES(evt event.ProductEvent) {
	ctx := context.Background()
	indexName := "products"
	productID := fmt.Sprintf("%d", evt.Product.ID)

	log.Printf("[ES-SYNC] Processing action %s for Product ID %s", evt.Action, productID)

	switch evt.Action {
	case event.ActionCreate, event.ActionUpdate:
		// 1. Prepare Data
		data, err := json.Marshal(evt.Product)
		if err != nil {
			log.Printf("[ERROR] Marshal JSON: %v", err)
			return
		}

		// 2. Indexing (Insert/Replace)
		// Menggunakan ID produk sebagai ID dokumen ES (Idempotent)
		res, err := h.esClient.Index(
			indexName,
			bytes.NewReader(data),
			h.esClient.Index.WithDocumentID(productID),
			h.esClient.Index.WithContext(ctx),
		)
		if err != nil {
			log.Printf("[ERROR] ES Indexing: %v", err)
			return
		}
		defer res.Body.Close()
		if res.IsError() {
			log.Printf("[ERROR] ES Indexing Response: %s", res.String())
		} else {
			log.Printf("✅ Product %s synced to Elasticsearch!", productID)
		}

	case event.ActionDelete:
		// Delete Document
		res, err := h.esClient.Delete(
			indexName,
			productID,
			h.esClient.Delete.WithContext(ctx),
		)
		if err != nil {
			log.Printf("[ERROR] ES Deleting: %v", err)
			return
		}
		defer res.Body.Close()
		// 404 Not Found saat delete itu wajar, abaikan
		if res.IsError() && res.StatusCode != 404 {
			log.Printf("[ERROR] ES Delete Response: %s", res.String())
		} else {
			log.Printf("🗑️ Product %s deleted from Elasticsearch!", productID)
		}
	}

	auditData := AuditLog{
		Timestamp: time.Now(),
		Action:    evt.Action,
		ProductID: evt.Product.ID,
		Payload:   evt.Product,
	}

	logBytes, _ := json.Marshal(auditData)

	res, err := h.esClient.Index(
		"product-logs", // Index baru!
		bytes.NewReader(logBytes),
		h.esClient.Index.WithContext(ctx),
	)

	if err == nil {
		defer res.Body.Close()
		log.Printf("📝 Audit Log %s recorded to Elasticsearch", evt.Action)
	}
}
