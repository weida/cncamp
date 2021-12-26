package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"math/rand"
	"time"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"runtime"
	"strings"
	"path/filepath"
	"io/ioutil"
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

type Service struct {
	Address string `yaml:"address"`
	Port    string `yaml:"port"`
}

type Log struct {
	Level string `yaml:"level"`
}

type Config struct {
	Service Service `yaml:"service"`
	Log     Log     `yaml:"log"`
}


func init() {
        log.SetFormatter(&log.TextFormatter{
           ForceColors:     false,
           DisableColors:   true,
           FullTimestamp:   true,
           TimestampFormat: "2006-01-02 15:04:05.000",
           CallerPrettyfier: func(frame *runtime.Frame) (function string, file string) {
               ss := strings.Split(frame.Function, ".")
               function = ss[len(ss)-1]
               file = fmt.Sprintf("%s:%d", filepath.Base(frame.File), frame.Line)
               return function, file
           },
        })
	log.SetOutput(os.Stdout)
	cfgFile, err := ioutil.ReadFile("/config/config.yaml")
	if err != nil {
	  panic (err)
	}
	cfg := new(Config)
	err = yaml.Unmarshal(cfgFile,cfg)
	level, err := log.ParseLevel(cfg.Log.Level)
	if err != nil {
	  panic(err)
	}
	log.SetLevel(level)
	listenAddr = cfg.Service.Address+":"+cfg.Service.Port
}

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
		log.Infof("-> [%s] %s %s %s %v\n", getRequestId(r), r.RemoteAddr, r.Method, r.URL, http.StatusNotFound)
		return
	}
	headers(w, r)
	fmt.Fprintf(w, "hello world in httpserver0\n")
	log.Infof("-> [%s] %s %s %s %v\n", getRequestId(r), r.RemoteAddr, r.Method, r.URL, http.StatusOK)
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
		requestID := r.Header.Get("X-Request-Id-Priv")
		if requestID == "" {
			requestID = nextRequestID()
		}
		w.Header().Set("X-Request-Id-Priv", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		log.Infof("<- [%s] %s %s %s \n", requestID, r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UTC().UnixNano())
	return min + rand.Intn(max-min)
}

func hello(w http.ResponseWriter, r *http.Request) {
	log.Debugf("< - entering hello handler")
	delay := randInt(10, 2000)
	time.Sleep(time.Millisecond*time.Duration(delay))
	fmt.Fprintf(w, "===================Details of the http request header:============\n")
	req, err := http.NewRequest("GET", "http://httpserver-service1/hello", nil)
	if  err != nil {
		log.Errorf("http NewRequest %s", err)
	}
	lowerCaseHeader := make(http.Header)
	for key, value := range r.Header {
		lowerCaseHeader[strings.ToLower(key)] = value
	}
	log.Infof("header:", lowerCaseHeader)
	req.Header = lowerCaseHeader
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("HTTP get failed with error:", "error", err)
	} else {
		log.Info("HTTP get succeeded")
	}
	if resp != nil {
		resp.Write(w)
	}
	log.Debugf("-> Respond %s ms delay", delay)
}


func main() {

	log.SetOutput(os.Stdout)
	log.Infoln("Server is starting...")

	http.HandleFunc("/", index)
	http.HandleFunc("/healthz", healthz)
	http.HandleFunc("/hello", hello)

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
		log.Infoln("Server is shutting...")
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
	log.Infoln("Server stopped")

}
