package main

import (
	"errors"
	"fmt"
	"greenlight/internal/data"
	"greenlight/internal/validator"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// authenticate checks if a Authorization token is given, and places a user into the Request.Context accordingly
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		//indica a cache que a resposta pode variar
		w.Header().Add("Vary", "Authorization")

		// retrieve the value from authorization header
		authorizationHeader := r.Header.Get("Authorization")

		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")

		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()
		data.ValidateTokenPlainText(v, token)
		if !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)

		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)

			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)

		next.ServeHTTP(w, r)
	})

}

// requireActivatedUser checks for anonymous users
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requireAuthenticatedUser checks for both authenticated and activated
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if !user.Ativated {
			app.incativeAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

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
	if !app.config.limiter.enable {
		return next
	}

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// launch a background dorotine which removes old entries from the clients map onc every minute
	go func() {
		for { // while true
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}

	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ip := realip.RealIP(r)

		mu.Lock()

		_, found := clients[ip]
		if !found {
			// 2 request per second, burst of 4 requests
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
		}

		clients[ip].lastSeen = time.Now()

		// when we call Allow, 1 token will be consumed
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceedResponse(w, r)
			return
		}

		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
