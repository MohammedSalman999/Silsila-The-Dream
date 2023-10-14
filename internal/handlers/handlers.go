package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/mohammedsalman999/silsila/internal/config"
	"github.com/mohammedsalman999/silsila/internal/forms"
	"github.com/mohammedsalman999/silsila/internal/helpers"
	"github.com/mohammedsalman999/silsila/internal/models"
	"github.com/mohammedsalman999/silsila/internal/render"
	"github.com/mohammedsalman999/silsila/internal/repository"
	"github.com/mohammedsalman999/silsila/internal/repository/dbrepo"
	"golang.org/x/crypto/bcrypt"
)

// Repo the repository used by the handlers
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig, db *sql.DB) *Repository {
	return &Repository{
		App: a,
		DB:  dbrepo.NewMysqlRepo(db, a),
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

// Home is the handler for the home page
func (m *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	m.App.Session.Put(r.Context(), "remote_ip", remoteIP)
	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

// About is the handler for the about page
func (m *Repository) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	stringMap := make(map[string]string)
	stringMap["test"] = "Hello, again"

	remoteIP := m.App.Session.GetString(r.Context(), "remote_ip")
	stringMap["remote_ip"] = remoteIP

	// send data to the template
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}

// Reservation renders the make a reservation page and displays form
func (m *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	var emptyReservation models.Reservation
	data := make(map[string]interface{})
	data["reservation"] = emptyReservation

	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

// PostReservation handles the posting of a reservation form
// PostReservation handles the posting of a reservation form
func (m *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		helpers.ServerError(w, err)
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, sd)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email")
	form.MinLength("first_name", 3, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	newReservationID, err := m.DB.InsertReservation(reservation)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	restriction := models.RoomRestriction{
		StartDate:     startDate,
		EndDate:       endDate,
		RoomID:        roomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
		Room:          models.Room{},
	}

	// Insert the room restriction
	err = m.DB.InsertRoomRestrictions(restriction)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	m.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)
}

// Generals renders the room page
func (m *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.tmpl", &models.TemplateData{})
}

// Majors renders the room page
func (m *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors.page.tmpl", &models.TemplateData{})
}

// Availability renders the search availability page
func (m *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

// Takes url parameter , and rediect user to the reservation section
func (m *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	sd := r.URL.Query().Get("s")
	ed := r.URL.Query().Get("e")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	var res models.Reservation

	res.RoomID = roomID
	res.StartDate = startDate
	res.EndDate = endDate

	m.App.Session.Put(r.Context(), "reservation", res)

	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

// PostAvailability handles post
func (m *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")

	// Format and print the start and end dates using fmt.Printf
	fmt.Printf("Start date is %s and end is %s\n", start, end)

	// Parse the dates
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, start)
	endDate, _ := time.Parse(layout, end)

	roomId, _ := strconv.Atoi(r.Form.Get("room_id"))

	// Replace this with your actual availability checking logic
	// SearchAvailabilityByRoomID should check if the room is available for the specified dates
	available, _ := m.DB.SearchAvailabilityByRoomID(startDate, endDate, roomId)

	if available {
		// If a room is available, redirect to the reservation page
		http.Redirect(w, r, "/make-reservation?id="+strconv.Itoa(roomId)+"&s="+start+"&e="+end, http.StatusSeeOther)
		return
	}

	resp := jsonResponse{
		OK:        false,
		Message:   "No rooms available for the selected dates",
		StartDate: start,
		EndDate:   end,
		RoomID:    strconv.Itoa(roomId),
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// AvailabilityJSON handles request for availability and sends JSON response
func (m *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {

	sd := r.Form.Get("start")
	ed := r.Form.Get("end")

	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, sd)
	endDate, _ := time.Parse(layout, ed)

	roomId, _ := strconv.Atoi(r.Form.Get("room_id"))

	available, _ := m.DB.SearchAvailabilityByRoomID(startDate, endDate, roomId)

	resp := jsonResponse{
		OK:        available,
		Message:   "Available!",
		StartDate: sd,
		EndDate:   ed,
		RoomID:    strconv.Itoa(roomId),
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

// Contact renders the contact page
func (m *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

// ReservationSummary displays the res summary page
func (m *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := m.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		log.Println("can't get item from session")
		m.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	m.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

// Use To Login Page

func (m *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
	})
}

// Logout handles the user logout process.
func (m *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	// Destroy the user's session.
	err := m.App.Session.Destroy(r.Context())
	if err != nil {
		// Handle the error if session destruction fails.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Renew the session token after destroying the session.
	err = m.App.Session.RenewToken(r.Context())
	if err != nil {
		// Handle the error if token renewal fails.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect the user to a relevant page after logout, e.g., the homepage.
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//sign-up page handler

func (m *Repository) ShowSignUp(w http.ResponseWriter, r *http.Request) {
	// Create an empty user object to pass to the template
	emptyUser := models.User{}

	// Create a map to hold the data to be passed to the template
	data := make(map[string]interface{})
	data["user"] = emptyUser

	// Render the signup page template and pass the empty user object
	render.RenderTemplate(w, r, "signup.page.tmpl", &models.TemplateData{
		Form: forms.New(nil),
		Data: data,
	})
}

//PostShowLogin in there

func (m *Repository) PostShowLogin(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	form := forms.New(r.PostForm)
	form.Required("email", "password")
	if !form.Valid() {
		render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	id, _, err := m.DB.Authenticate(email, password)
	if err != nil {
		log.Println(err)

		m.App.Session.Put(r.Context(), "error", "invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	m.App.Session.Put(r.Context(), "user_id", id)
	// Set a session variable to indicate successful login.
	m.App.Session.Put(r.Context(), "login_success", true)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// PostShowSignup handles user registration through a form.
func (m *Repository) PostShowSignup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		helpers.ServerError(w, err)
		return
	}

	user := models.User{
		FirstName:   r.Form.Get("first_name"),
		LastName:    r.Form.Get("last_name"),
		Email:       r.Form.Get("email"),
		Password:    r.Form.Get("password"), // Assuming you have a field for password in your form
		AccessLevel: 1,                      // You can set the access level as needed
	}

	form := forms.New(r.PostForm)
	form.Required("first_name", "last_name", "email", "password")
	form.MinLength("first_name", 3, r)
	form.IsEmail("email")

	if !form.Valid() {
		data := make(map[string]interface{})
		data["user"] = user
		render.RenderTemplate(w, r, "signup.page.tmpl", &models.TemplateData{
			Form: form,
			Data: data,
		})
		return
	}

	// Use the standard password hashing function
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	user.Password = hashedPassword

	// Insert the user into the database
	newUserID, err := m.DB.InsertUser(user)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	// Optionally, you can do further processing or redirection after user insertion.
	// For example, you can set a session variable and redirect to a user profile page.

	// Set a session variable (if you are using sessions)
	m.App.Session.Put(r.Context(), "user_id", newUserID)

	// Redirect to a user profile page or another relevant page
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

//Social media login

func (m *Repository) SocialLogin(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	m.App.Session.Put(r.Context(), "social_provider", provider)
	m.InItSocialAuth(provider) // Call InItSocialAuth with the provider parameter

	if _, err := gothic.CompleteUserAuth(w, r); err == nil {
		// User is already logged in
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		// Attempt social login
		gothic.BeginAuthHandler(w, r)
	}
}

func (m *Repository) InItSocialAuth(provider string) {
	switch provider {
	case "google":
		goth.UseProviders(
			google.New(
				"523876759739-f51u9b231dne99aq2lcjdofdn2jcfpop.apps.googleusercontent.com",
				"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",
				"http://localhost:3000/auth/google/callback",
				"email",   // Requesting access to the user's email
				"profile", // Requesting access to the user's profile information
			),
		)
	// Add more providers as needed

	default:
		// Handle other providers or show an error if needed
	}

	key := os.Getenv("KEY")
	maxAge := 86400 * 30

	st := sessions.NewCookieStore([]byte(key))
	st.MaxAge(maxAge)
	st.Options.Path = "/"
	st.Options.HttpOnly = true
	st.Options.Secure = false

	// Add your session store initialization code here if needed.

	gothic.Store = st

	// ...
}

// SocialLoginCallback handles the Google OAuth callback and user registration.
func (m *Repository) SocialLoginCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	m.InItSocialAuth("google") // Initialize social authentication for Google

	gUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		m.App.Session.Put(r.Context(), "error", err.Error())
		http.Redirect(w, r, "/users/login", http.StatusSeeOther)
		return
	}

	// Generate a random password for Google-authenticated users
	password := GenerateRandomPassword(13)

	// Use the standard password hashing function
	hashedPassword, err := HashPassword(password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newUser := models.User{
		FirstName:   gUser.FirstName,
		LastName:    gUser.LastName,
		Email:       gUser.Email,
		Password:    hashedPassword,
		AccessLevel: 1,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Insert the user into the database
	_, err = m.DB.InsertUser(newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Continue with your desired logic after handling user insertion
	// ...

	// Redirect to the home page or any other relevant page on success
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// GenerateRandomPassword generates a random password of a specified length.
func GenerateRandomPassword(length int) string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Use a secure random source
	source := rand.NewSource(time.Now().UnixNano())
	generator := rand.New(source)

	password := make([]byte, length)
	for i := 0; i < length; i++ {
		randomIndex := generator.Intn(len(charset))
		password[i] = charset[randomIndex]
	}
	return string(password)
}

// HashPassword hashes a plain text password using bcrypt.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}
