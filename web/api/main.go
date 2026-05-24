package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codeatlasdev/domain-hunter/web/api/handlers"
	"github.com/codeatlasdev/domain-hunter/web/api/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	checkLimiter := middleware.NewRateLimiter(0.5, 30)    // 30 req/min burst
	generalLimiter := middleware.NewRateLimiter(1.67, 100) // 100 req/min burst

	mux := http.NewServeMux()
	mux.Handle("POST /api/check", checkLimiter.Wrap(http.HandlerFunc(handlers.HandleCheck)))
	mux.Handle("GET /api/check/stream", checkLimiter.Wrap(http.HandlerFunc(handlers.HandleCheckStream)))
	mux.Handle("GET /api/presets", generalLimiter.Wrap(http.HandlerFunc(handlers.HandlePresets)))
	mux.Handle("GET /api/tlds", generalLimiter.Wrap(http.HandlerFunc(handlers.HandleTLDs)))
	mux.Handle("GET /api/prices/", generalLimiter.Wrap(http.HandlerFunc(handlers.HandlePrices)))

	handler := cors(mux)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		fmt.Printf("domh-api listening on :%s\n", port)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	fmt.Println("server stopped")
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
