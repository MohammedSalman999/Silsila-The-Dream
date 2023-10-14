package main

import (
	"net/http"

	"github.com/justinas/nosurf"
	"github.com/mohammedsalman999/silsila/internal/helpers"
)

// NoSurf is the csrf protection middleware
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}

// SessionLoad loads and saves session data for current request
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}


func Auth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Your authentication logic goes here

        // For example, you can check if the user is authenticated
        if !helpers.IsAuthenticated(r) {
            // Set an error message in the session
            session.Put(r.Context(), "error", "Login First")
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // Call the next handler in the chain
        next.ServeHTTP(w, r)
    })
}

