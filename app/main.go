package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// build trigger: 1

type job struct {
	ID      string          `json:"id"`
	Webhook string          `json:"webhook"`
	Payload json.RawMessage `json:"payload"`
}

type result struct {
	ID      string          `json:"id"`
	Status  string          `json:"status"`
	Payload json.RawMessage `json:"payload"`
}

const (
	queueSize       = 100
	workerCount     = 4
	shutdownTimeout = 10 * time.Second
	requestTimeout  = 5 * time.Second
)

var jobCounter int64
var commit = "unknown"

type jobHandler struct {
	queue chan<- job
}

func newJobHandler(queue chan<- job) *jobHandler {
	return &jobHandler{queue: queue}
}

func (h *jobHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var j job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil || j.Webhook == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	j.ID = fmt.Sprintf("%d", atomic.AddInt64(&jobCounter, 1))

	select {
	case h.queue <- j:
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"id": j.ID})
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

type webhookHandler struct {
	client *http.Client
}

func newWebhookHandler() *webhookHandler {
	return &webhookHandler{client: &http.Client{Timeout: requestTimeout}}
}

func (h *webhookHandler) handle(j job) {
	time.Sleep(time.Duration(200+rand.Intn(1300)) * time.Millisecond)
	body, err := json.Marshal(result{ID: j.ID, Status: "done", Payload: j.Payload})
	if err != nil {
		return
	}
	resp, err := h.client.Post(j.Webhook, "application/json", bytes.NewReader(body))
	if err == nil {
		resp.Body.Close()
	}
}

func worker(queue <-chan job, wg *sync.WaitGroup, wh *webhookHandler) {
	defer wg.Done()
	for j := range queue {
		wh.handle(j)
	}
}

type healthHandler struct {
	ready  *atomic.Bool
	commit string
}

func newHealthHandler(ready *atomic.Bool, commit string) *healthHandler {
	return &healthHandler{ready: ready, commit: commit}
}

func (h *healthHandler) livez(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"commit": h.commit})
}

func (h *healthHandler) readyz(w http.ResponseWriter, r *http.Request) {
	if !h.ready.Load() {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	log.Print("application starting up...")
	queue := make(chan job, queueSize)
	var wg sync.WaitGroup
	wh := newWebhookHandler()

	for range workerCount {
		wg.Add(1)
		go worker(queue, &wg, wh)
	}

	var ready atomic.Bool
	ready.Store(true)
	hh := newHealthHandler(&ready, commit)

	mux := http.NewServeMux()
	mux.Handle("/jobs", newJobHandler(queue))
	mux.HandleFunc("/livez", hh.livez)
	mux.HandleFunc("/readyz", hh.readyz)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{Addr: ":" + port, Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	log.Print("application startup completed")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Print("application shutting down...")

	ready.Store(false)

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	srv.Shutdown(ctx)

	close(queue)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
	}

	log.Print("application shut down complete")
}
