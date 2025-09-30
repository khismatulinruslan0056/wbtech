package middleware

import (
	"L0/internal/transport/httpserver/common"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestValidateOrderID(t *testing.T) {
	var (
		id     string
		status int
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Context().Value(common.OrderIDKey).(string); got != id && got != "" {
			t.Errorf("orderID got %v, want %v", got, id)
		}
		w.WriteHeader(status)
	})

	testcases := []struct {
		name   string
		called bool
		id     string
		status int
	}{
		{
			name:   "valid1",
			called: true,
			id:     "Abcdef123456",
			status: http.StatusNoContent,
		},
		{
			name:   "valid2",
			called: true,
			id:     strings.Repeat("a", 64),
			status: http.StatusOK,
		},
		{
			name:   "invalid",
			called: false,
			id:     "",
			status: http.StatusBadRequest,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			id = tc.id
			status = tc.status
			h := ValidateOrderID(next)

			req := httptest.NewRequest(http.MethodGet, "/order/"+id, nil)
			req.SetPathValue("orderID", tc.id)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)
			if rr.Code != tc.status {
				t.Errorf("status: got %v, want %v", rr.Code, tc.status)
			}
			if !tc.called {
				ct := rr.Header().Get("Content-Type")
				if !strings.HasPrefix(ct, "application/json") {
					t.Fatalf("Content-Type = %q, want application/json", ct)
				}
				if !strings.Contains(rr.Body.String(), "invalid or empty orderID") {
					t.Fatalf("body %q does not contain error message", rr.Body.String())
				}
			}
		})
	}
}

type capturedRecord struct {
	Level   slog.Level
	Message string
	Attrs   map[string]any
}

type memHandler struct {
	records []capturedRecord
}

func (h *memHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *memHandler) Handle(ctx context.Context, r slog.Record) error {
	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})
	h.records = append(h.records, capturedRecord{
		Level:   r.Level,
		Message: r.Message,
		Attrs:   attrs,
	})
	return nil
}
func (h *memHandler) WithAttrs(as []slog.Attr) slog.Handler { return h }
func (h *memHandler) WithGroup(name string) slog.Handler    { return h }

func newTestLogger() (*slog.Logger, *memHandler) {
	h := &memHandler{}
	return slog.New(h), h
}

func TestLogger(t *testing.T) {
	var (
		status      int
		msg         string
		writeHeader bool
		expectedErr bool
	)
	log, mh := newTestLogger()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if writeHeader {
			w.WriteHeader(status)
		}
		if expectedErr {
			_, _ = io.WriteString(w, msg)
		} else {
			_, _ = w.Write([]byte(msg))
		}
	})

	testcases := []struct {
		name        string
		status      int
		msg         string
		writeHeader bool
		expectedErr bool
		path        string
		level       slog.Level
	}{
		{
			name:        "valid1",
			status:      http.StatusOK,
			msg:         "ok",
			writeHeader: true,
			expectedErr: false,
			path:        "/x",
			level:       slog.LevelInfo,
		},
		{
			name:        "invalid",
			status:      http.StatusBadRequest,
			msg:         strings.Repeat("X", 600),
			writeHeader: true,
			expectedErr: true,
			path:        "/err",
			level:       slog.LevelError,
		},
		{
			name:        "valid2",
			status:      http.StatusOK,
			msg:         "payload",
			writeHeader: false,
			expectedErr: false,
			path:        "/asd",
			level:       slog.LevelInfo,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			status = tc.status
			msg = tc.msg
			writeHeader = tc.writeHeader
			expectedErr = tc.expectedErr
			h := Logger(log)(next)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tc.status {
				t.Fatalf("status = %d, want %d", rr.Code, tc.status)
			}
			if len(mh.records) < 1 {
				t.Fatalf("log records = %d, must be more then 1", len(mh.records))
			}
			rec := mh.records[len(mh.records)-1]
			if rec.Level != tc.level {
				t.Fatalf("log level = %v, want %v", rec.Level, tc.level)
			}
			if rec.Attrs["Method"] != http.MethodGet {
				t.Fatalf("Method = %v, want GET", rec.Attrs["Method"])
			}
			if rec.Attrs["Path"] != tc.path {
				t.Fatalf("Path = %v, want %s", rec.Attrs["Path"], tc.path)
			}

			if got := rec.Attrs["statusCode"]; got != int64(tc.status) {
				t.Fatalf("statusCode = %v, want %d", got, tc.status)
			}
			if _, ok := rec.Attrs["duration"].(time.Duration); !ok {
				t.Fatalf("duration attr not present or wrong type: %T", rec.Attrs["duration"])
			}

			if tc.expectedErr {
				errMsg, _ := rec.Attrs["errMsg"].(string)
				if len(errMsg) == 0 {
					t.Fatal("errMsg is empty, want truncated body")
				}
				if !strings.HasSuffix(errMsg, "...") {
					t.Fatalf("errMsg must be truncated and end with '...'; got: %q", msg)
				}
				if len(errMsg) < 512 {
					t.Fatalf("errMsg length = %d, want >= 512", len(msg))
				}
			}
		})
	}

}

func TestRecoverPanic(t *testing.T) {
	log, mh := newTestLogger()
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("oops")
	})

	h := Logger(log)(RecoverPanic(log)(next))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}

	if len(mh.records) < 2 {
		t.Fatalf("log records = %d, must be more then 2", len(mh.records))
	}

	rec1 := mh.records[len(mh.records)-2]
	rec2 := mh.records[len(mh.records)-1]

	if rec1.Message != "panic" || rec2.Level != slog.LevelError {
		t.Fatalf("first record must be panic error, got: %+v", rec1)
	}
	if code := rec2.Attrs["statusCode"]; code != int64(http.StatusInternalServerError) {
		t.Fatalf("second record statusCode = %v, want 500", code)
	}
	if rec2.Level != slog.LevelError {
		t.Fatalf("second record level = %v, want ERROR", rec2.Level)
	}
}
