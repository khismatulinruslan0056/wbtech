package router

import (
	"L0/internal/service"
	"L0/internal/transport/httpserver/handlers"
	"L0/internal/transport/httpserver/middleware"
	"log/slog"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Router struct {
	mux *http.ServeMux
}

func New(s *service.Service, log *slog.Logger) *Router {
	logMiddleware := middleware.Logger(log)
	recoverMiddleware := middleware.RecoverPanic(log)

	getOrderHandler := handlers.GetOrder(log, s)

	mux := http.NewServeMux()

	mux.Handle("GET /order/{orderID}",
		logMiddleware(
			recoverMiddleware(
				middleware.ValidateOrderID(getOrderHandler),
			),
		),
	)

	mux.Handle("GET /ping",
		logMiddleware(
			recoverMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("pong"))
				}),
			),
		),
	)

	mux.Handle("GET /healthz",
		logMiddleware(
			recoverMiddleware(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("ok"))
				}),
			),
		),
	)

	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	fileServer := http.FileServer(http.Dir("./web"))
	mux.Handle("/", fileServer)

	r := &Router{
		mux: mux,
	}
	return r
}

func (rout *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rout.mux.ServeHTTP(w, r)
}
