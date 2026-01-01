package search

import (
	"log"

	"github.com/elastic/go-elasticsearch/v7"
)

func InitES(address string) *elasticsearch.Client {
	cfg := elasticsearch.Config{
		Addresses: []string{address},
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatalf("Error creaating the client: %s", err)
	}

	res, err := es.Info()
	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}
	defer res.Body.Close()

	log.Println("✅ Terhubung ke Elasticsearch!")
	return es
}
