package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
)

// Setup Logger Global (JSON Format)
func InitLogger() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 1. Generate a unique request ID
		reqID := uuid.New().String()

		// 2. Wrap Writer
		wrappedWriter := NewResponseWriterWrapper(w)

		// 3. Tambah Header X-Request-ID (Biar Frontend Bisa Baca)
		w.Header().Set("X-Request-ID", reqID)

		// 4. Lanjut ke Handler asli
		next.ServeHTTP(wrappedWriter, r)

		// 5. Hitung durasi dan log
		duration := time.Since(start)

		// 6. Log menggunakan slog
		slog.Info("incoming request",
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", wrappedWriter.statusCode),
			slog.String("duration", duration.String()),
			slog.String("ip", r.RemoteAddr),
		)
	})
}
