package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/mohammedsalman999/silsila/internal/config"
	"github.com/mohammedsalman999/silsila/internal/handlers"

	"net/http"
)

func routes(app *config.AppConfig) http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(NoSurf)
	mux.Use(SessionLoad)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)
	mux.Get("/generals-quarters", handlers.Repo.Generals)
	mux.Get("/majors-suite", handlers.Repo.Majors)

	mux.Get("/search-availability", handlers.Repo.Availability)
	mux.Post("/search-availability", handlers.Repo.PostAvailability)
	mux.Post("/search-availability-json", handlers.Repo.AvailabilityJSON)
	mux.Get("/book-room", handlers.Repo.BookRoom)

	mux.Get("/contact", handlers.Repo.Contact)

	mux.Get("/make-reservation", handlers.Repo.Reservation)
	mux.Post("/make-reservation", handlers.Repo.PostReservation)
	mux.Get("/reservation-summary", handlers.Repo.ReservationSummary)

	//User Login Page
	mux.Get("/user/login", handlers.Repo.ShowLogin)      // GET method
	mux.Post("/user/login", handlers.Repo.PostShowLogin) // POST method
	mux.Get("/user/logout", handlers.Repo.Logout)// Logout 


	//Sign-Up Page 
	mux.Get("/user/signup", handlers.Repo.ShowSignUp) //get method
	mux.Post("/user/signup", handlers.Repo.PostShowSignup)


	//Social login auth 

	mux.Get("/auth/{provider}", handlers.Repo.SocialLogin)
	mux.Get("/auth/{provider}/callback", handlers.Repo.SocialLoginCallback)



	fileServer := http.FileServer(http.Dir("./static/"))
	mux.Handle("/static/*", http.StripPrefix("/static", fileServer))

	return mux
}
