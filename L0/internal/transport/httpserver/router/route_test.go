package router

import (
	"L0/internal/service"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func NewTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	var s service.Service
	r := New(&s, slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return httptest.NewServer(r)
}

func TestRoute(t *testing.T) {
	ts := NewTestServer(t)
	defer ts.Close()
	testCases := []struct {
		name       string
		method     string
		path       string
		wantCode   int
		wantBody   string
		wantJSONCT bool
		wantAllow  string
	}{
		{
			name:     "ping",
			method:   http.MethodGet,
			wantCode: http.StatusOK,
			path:     "/ping",
			wantBody: "pong",
		},
		{
			name:     "healthz",
			method:   http.MethodGet,
			wantCode: http.StatusOK,
			path:     "/healthz",
			wantBody: "ok",
		},
		{
			name:       "bad request",
			method:     http.MethodGet,
			wantCode:   http.StatusBadRequest,
			path:       "/order/bad",
			wantJSONCT: true,
		},
		{
			name:     "swagger",
			method:   http.MethodGet,
			wantCode: http.StatusOK,
			path:     "/swagger/",
		},
		{
			name:     "nope",
			method:   http.MethodGet,
			wantCode: http.StatusNotFound,
			path:     "/nope",
		},
	}

	client := http.DefaultClient
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, ts.URL+tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("%s %s, error %v", tc.method, tc.path, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.wantCode {
				t.Errorf("expected status code %d, got %d", tc.wantCode, resp.StatusCode)
			}

			if tc.wantJSONCT {
				ct := resp.Header.Get("Content-Type")
				if !strings.HasPrefix(ct, "application/json") {
					t.Fatalf("Content-Type = %q, want application/json", ct)
				}
			}

			if tc.wantAllow != "" {
				if got := resp.Header.Get("Allow"); got != tc.wantAllow {
					t.Fatalf("Allow = %q, want %q", got, tc.wantAllow)
				}
			}

			if tc.wantBody != "" {
				b, _ := io.ReadAll(resp.Body)
				if string(b) != tc.wantBody {
					t.Fatalf("body = %q, want %q", string(b), tc.wantBody)
				}
			}
		})
	}
}
