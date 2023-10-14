package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/alexedwards/scs/v2"
	"github.com/mohammedsalman999/silsila/internal/config"
	"github.com/mohammedsalman999/silsila/internal/handlers"
	"github.com/mohammedsalman999/silsila/internal/models"
	"github.com/mohammedsalman999/silsila/internal/render"

	"log"
	"net/http"
	"time"
)

const portNumber = ":3000"

var app config.AppConfig
var session *scs.SessionManager

// main is the main function
func main() {
	// what am I going to put in the session
	gob.Register(models.Reservation{})
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.RoomRestriction{})

	// change this to true when in production
	app.InProduction = false

	// set up the session
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = app.InProduction

	app.Session = session

	// connect to databae

	log.Println("Connecting to database...")
	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/silsila")
	if err != nil {
		panic(err)
	}
	defer db.Close() // this will be executed when the main function will exit

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to the database!")

	// Perform database operations here...

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create template cache")
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, db)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)

	fmt.Println(fmt.Sprintf("Staring application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
