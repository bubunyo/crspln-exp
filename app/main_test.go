package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestJobHandlerAccepted(t *testing.T) {
	queue := make(chan job, 1)
	h := newJobHandler(queue)
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"webhook":"http://example.com","payload":{"x":1}}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 got %d", rec.Code)
	}

	select {
	case j := <-queue:
		if j.Webhook != "http://example.com" {
			t.Fatalf("unexpected webhook %q", j.Webhook)
		}
	default:
		t.Fatal("job not queued")
	}
}

func TestJobHandlerBadRequest(t *testing.T) {
	queue := make(chan job, 1)
	h := newJobHandler(queue)
	cases := []string{
		`not json`,
		`{"payload":{"x":1}}`,
	}
	for _, c := range cases {
		req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(c))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d for %q", rec.Code, c)
		}
	}
}

func TestJobHandlerMethodNotAllowed(t *testing.T) {
	queue := make(chan job, 1)
	h := newJobHandler(queue)
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 got %d", rec.Code)
	}
}

func TestJobHandlerQueueFull(t *testing.T) {
	queue := make(chan job, 1)
	queue <- job{ID: "x"}
	h := newJobHandler(queue)
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"webhook":"http://example.com"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 got %d", rec.Code)
	}
}

func TestWebhookHandlerHandle(t *testing.T) {
	received := make(chan result, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var res result
		json.NewDecoder(r.Body).Decode(&res)
		received <- res
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	wh := newWebhookHandler()
	wh.handle(job{ID: "1", Webhook: srv.URL, Payload: json.RawMessage(`{"x":1}`)})

	select {
	case res := <-received:
		if res.ID != "1" || res.Status != "done" {
			t.Fatalf("unexpected result %+v", res)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("webhook not called")
	}
}
