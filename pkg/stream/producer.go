package stream

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokers []string) *KafkaProducer {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll // tunggu sampai kafka benar benar simpan data (durability)
	config.Producer.Retry.Max = 5                    // retry jika network kumat

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Kafka producer: %v", err)
	}

	return &KafkaProducer{producer: producer}
}

// mengirim event ke topic tertentu
func (k *KafkaProducer) SendMessage(topic string, key string, message interface{}) error {
	val, err := json.Marshal(message)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key), // Key menentukan partition
		Value: sarama.ByteEncoder(val),
	}

	// mengirim partition
	partition, offset, err := k.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("[KAFKA] Message sent to topic %s | Partition: %d | Offset: %d", topic, partition, offset)
	return nil
}

func (k *KafkaProducer) Close() {
	k.producer.Close()
}
