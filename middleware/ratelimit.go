package middleware

import (
	"net"
	"net/http"
	"phase3-api-architecture/utils"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Menyimpan map limiter untuk setiap IP
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	// Background cleanup untuk menghapus IP lama agar RAM tidak bocor
	go i.cleanupVisitors()

	return i
}

// mengambil atau membuat limiter untuk ip tertentu
func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.ips[ip]
	if !exists {
		limiter = rate.NewLimiter(i.r, i.b)
		i.ips[ip] = limiter
	}

	return limiter
}

// menghapus data ip setiap 5 menit
func (i *IPRateLimiter) cleanupVisitors() {
	for {
		time.Sleep(5 * time.Minute)
		i.mu.Lock()
		// reset map
		i.ips = make(map[string]*rate.Limiter)
		i.mu.Unlock()
	}
}

// middleware func
func (i *IPRateLimiter) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ambil ip address asli user
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// jika gagal parse ip, anggap valid atau log error
			next.ServeHTTP(w, r)
			return
		}

		limiter := i.getLimiter(ip)

		// cek apakah boleh lewat
		if !limiter.Allow() {
			// REJECT! 429 TMR
			utils.ResponseError(w, http.StatusTooManyRequests, "Terlalu banyak request, santai dulu kawan")
			return
		}

		next.ServeHTTP(w, r)
	})
}
