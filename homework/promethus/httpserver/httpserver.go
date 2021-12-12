package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"httpserver/metrics"
	"math/rand"
)

type key int

const (
	requestIDKey key = 0
)

const (
	UP   int32 = 200
	DOWN int32 = 500
)

var (
	listenAddr string
	healthy    int32
)

func getRequestId(r *http.Request) string {
	requestID, ok := r.Context().Value(requestIDKey).(string)
	if !ok {
		requestID = "unknown"
	}
	return requestID
}

func writeback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Version", version())
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusNotFound)
		log.Printf("-> [%s] %s %s %s %v\n", getRequestId(r), r.RemoteAddr, r.Method, r.URL, http.StatusNotFound)
		return
	}
	headers(w, r)
	fmt.Fprintf(w, "hello world\n")
	log.Printf("-> [%s] %s %s %s %v\n", getRequestId(r), r.RemoteAddr, r.Method, r.URL, http.StatusOK)
}

func headers(w http.ResponseWriter, r *http.Request) {
	for name, headers := range r.Header {
		for _, h := range headers {
			w.Header().Set(name, h)
		}
	}
	writeback(w, r)
}

func healthz(w http.ResponseWriter, r *http.Request) {
	headers(w, r)

	if atomic.LoadInt32(&healthy) == UP || atomic.LoadInt32(&healthy) == DOWN {
		fmt.Fprintf(w, "%v", healthy)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
	log.Printf("-> [%s] %s %s %s %v\n", getRequestId(r), r.RemoteAddr, r.Method, r.URL, healthy)
}

func version() string {
	v := os.Getenv("VERSION")
	if v != "" {
		return v
	} else {
		return "Unknown"
	}
}
func nextRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func logrequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-Id")
		if requestID == "" {
			requestID = nextRequestID()
		}
		w.Header().Set("X-Request-Id", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		log.Printf("<- [%s] %s %s %s \n", requestID, r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Printf("< - entering hello handler")
	timer := metrics.NewTimer()
	defer timer.ObserveTotal()
	user := r.URL.Query().Get("user")
	delay := randInt(10,2000)
	time.Sleep(time.Millisecond*time.Duration(delay))
	if user != "" {
		fmt.Fprintf(w, "hello [%s]\n", user)
	} else {
		fmt.Fprintf(w, "hello [stranger]\n")
	}
	fmt.Fprintf(w, "===================Details of the http request header:============\n")
	for k, v := range r.Header {
		fmt.Fprintf(w, "%s = %s \n", k, v )
	}
	log.Printf("-> Respond in %d ms", delay)
}


func main() {

	flag.StringVar(&listenAddr, "listenAddr", ":8090", "port")
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.Println("Server is starting...")

	metrics.Register()

	http.HandleFunc("/", index)
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/hello", hello)
	http.Handle("/metrics", promhttp.Handler())

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGUSR1)

	server := &http.Server{
		Addr:         listenAddr,
		Handler:      logrequest(http.DefaultServeMux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		<-quit
		log.Println("Server is shutting...")
		atomic.StoreInt32(&healthy, DOWN)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		server.SetKeepAlivesEnabled(false)

		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shutdown the server %v\n", err)
		}
		close(done)
	}()

	atomic.StoreInt32(&healthy, UP)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen no %s : %v\n", listenAddr, err)
	}

	<-done
	log.Println("Server stopped")

}
