package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	defaultPort          = 8080
	traceparentHeaderKey = "traceparent"
)

type Response struct {
	Service     string `json:"service"`
	Message     string `json:"message"`
	At          string `json:"at"`
	Traceparent string `json:"traceparent"`
}

func main() {
	var (
		ServiceName          = ""
		DownstreamServiceURL = ""
	)

	port := defaultPort
	if p := os.Getenv("PORT"); p != "" {
		var parsed int
		_, err := fmt.Sscanf(p, "%d", &parsed)
		if err == nil {
			port = parsed
		}
	}

	if s := os.Getenv("SERVICE_NAME"); s != "" {
		ServiceName = s
	} else {
		panic("SERVICE_NAME is not set")
	}

	if d := os.Getenv("DOWNSTREAM_SERVICE_URL"); d != "" {
		DownstreamServiceURL = d
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/call", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		traceparent := r.Header.Get(traceparentHeaderKey)

		log.Printf("[%s] incoming request %s %s traceparent=%s", ServiceName, r.Method, r.URL.Path, traceparent)
		log.Printf("[%s] request headers: %v", ServiceName, r.Header)

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		var resp *Response
		var err error

		if DownstreamServiceURL != "" {
			resp, err = call(ctx, DownstreamServiceURL)
			if err != nil {
				log.Printf("[%s] error calling DownstreamServiceURL %s: %v traceparent=%s", ServiceName, DownstreamServiceURL, err, traceparent)
				http.Error(w, fmt.Sprintf("%s calling DownstreamServiceURL %s: %v", err, ServiceName, DownstreamServiceURL), http.StatusBadGateway)
				return
			}
		} else {
			log.Printf("[%s] no downstream service configured, skipping call", ServiceName)
			resp = &Response{
				Service:     ServiceName,
				Message:     "no downstream service",
				At:          time.Now().Format(time.RFC3339Nano),
				Traceparent: traceparent,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set(traceparentHeaderKey, traceparent)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("[%s] encode response error: %v traceparent=%s", ServiceName, err, traceparent)
		}

		log.Printf("[%s] completed %s in %s traceparent=%s", ServiceName, DownstreamServiceURL, time.Since(start), traceparent)
	})

	addr := fmt.Sprintf(":%d", port)
	log.Printf("[%s] listening on %s", ServiceName, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[%s] server error: %v", ServiceName, err)
	}
}

func call(ctx context.Context, downstreamServiceURL string) (*Response, error) {
	url := downstreamServiceURL + "/call"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 2 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("downstream service status %d: %s", res.StatusCode, string(body))
	}

	var out Response
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}
