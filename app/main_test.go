package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestJobHandlerAccepted(t *testing.T) {
	queue := make(chan job, 1)
	h := newJobHandler(queue)
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"webhook":"http://example.com","payload":{"x":1}}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusAccepted, rec.Code)

	select {
	case j := <-queue:
		require.Equal(t, "http://example.com", j.Webhook)
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
		require.Equal(t, http.StatusBadRequest, rec.Code, "case %q", c)
	}
}

func TestJobHandlerMethodNotAllowed(t *testing.T) {
	queue := make(chan job, 1)
	h := newJobHandler(queue)
	req := httptest.NewRequest(http.MethodGet, "/jobs", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusMethodNotAllowed, rec.Code)
}

func TestJobHandlerQueueFull(t *testing.T) {
	queue := make(chan job, 1)
	queue <- job{ID: "x"}
	h := newJobHandler(queue)
	req := httptest.NewRequest(http.MethodPost, "/jobs", strings.NewReader(`{"webhook":"http://example.com"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestHealthHandlerLivez(t *testing.T) {
	var ready atomic.Bool
	h := newHealthHandler(&ready, "abc123")
	rec := httptest.NewRecorder()
	h.livez(rec, httptest.NewRequest(http.MethodGet, "/livez", nil))

	require.Equal(t, http.StatusOK, rec.Code)

	var body map[string]string
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
	require.Equal(t, "abc123", body["commit"])
}

func TestHealthHandlerReadyz(t *testing.T) {
	var ready atomic.Bool
	h := newHealthHandler(&ready, "abc123")

	rec := httptest.NewRecorder()
	h.readyz(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	ready.Store(true)
	rec = httptest.NewRecorder()
	h.readyz(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	require.Equal(t, http.StatusOK, rec.Code)
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
		require.Equal(t, "1", res.ID)
		require.Equal(t, "done", res.Status)
	case <-time.After(3 * time.Second):
		t.Fatal("webhook not called")
	}
}
