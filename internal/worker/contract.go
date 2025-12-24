package worker

type TaskSendInvoice struct {
	UserID     int    `json:"user_id"`
	Email      string `json:"email"`
	ProductID  int    `json:"product_id"`
	TotalPrice int    `json:"total_price"`
}

const QueueInvoice = `queue:invoice_sending`
