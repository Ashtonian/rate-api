package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type statusInterceptor struct {
	http.ResponseWriter
	status int
}

func (interceptor *statusInterceptor) WriteHeader(code int) {
	interceptor.status = code
	interceptor.ResponseWriter.WriteHeader(code)
}

type Middleware func(http.Handler) http.Handler

func MiddlewareChain(h http.Handler, m ...Middleware) http.Handler {
	if len(m) < 1 {
		return h
	}
	wrappedH := h
	for i := len(m) - 1; i >= 0; i-- {
		wrappedH = m[i](wrappedH)
	}

	return wrappedH
}

func NewMetricsMiddleware(store *MetricsStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wi := statusInterceptor{
				ResponseWriter: w,
				status:         http.StatusOK,
			}
			next.ServeHTTP(&wi, r)
			duration := time.Since(start).Milliseconds()
			if duration < 1 {
				duration = 1
			}

			store.Record(r.Method, r.URL.Path, wi.status, int(duration))
		})
	}
}

func NewRecoveryMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := recover()
				if err != nil {
					fmt.Printf("Api Panic Recovered:%s", err)
					response := ErrorResponse{
						Error: ErrInternal,
					}

					bod, _ := json.Marshal(response)

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write(bod)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
