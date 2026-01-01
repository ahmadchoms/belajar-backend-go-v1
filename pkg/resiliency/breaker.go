package resiliency

import (
	"errors"
	"log"
	"time"

	"github.com/sony/gobreaker"
)

var (
	ErrServiceUnavailbale = errors.New("service temporarily unavailable (circuit open)")
)

// NewDatabaseBreaker membuat settingan circuit breaker khusus untuk database
func NewDatabaseBreaker(name string) *gobreaker.CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        name,
		MaxRequests: 1,                // jumlah request yang boleh lewat saat fase "Half-Open" (percobaan)
		Interval:    60 * time.Second, // reset hitungan error setiap 10 detik
		Timeout:     30 * time.Second, // durasi sirkuit terbuka (mati) sebelum mencoba lagi

		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// jika rasio kegagalan > 60% dari minimal 5 request
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && counts.TotalFailures >= 3 && failureRatio >= 0.6
		},

		// callback untuk log
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("Circuit Breaker %s changed from %s to %s", name, from, to)
		},
	}
	return gobreaker.NewCircuitBreaker(settings)
}
