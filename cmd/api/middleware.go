package main

import (
	"fmt"
	"net/http"

	"golang.org/x/time/rate"
)

// this will only reover panics that happen in this goroutine in case someone starts another process
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		defer func() {
			// checks if a panic accors
			pv := recover()
			if pv != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%v", pv))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {

	// 2 request per second,
	// burst of 4 requests
	limiter := rate.NewLimiter(2, 4)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// when we call Allow, 1 token will be consumed
		if !limiter.Allow() {
			app.rateLimitExceedResponse(w, r)
			return
		}
		/// codigo
		next.ServeHTTP(w, r)
	})
}
