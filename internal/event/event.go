package event

import "phase3-api-architecture/models"

// Constants untuk tipe akse
const (
	ActionCreate = "CREATE"
	ActionUpdate = "UPDATE"
	ActionDelete = "DELETE"
)

// payload yang dikirim ke kafka topic 'product-events'
type ProductEvent struct {
	Action  string         `json:"action"`
	Product models.Product `json:"payload"`
}
