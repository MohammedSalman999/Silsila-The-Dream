package models

import "time"

// User Is the Users Model
type User struct {
	ID          int
	FirstName   string
	LastName    string
	Email       string
	Password    string
	AccessLevel int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Room is the rooms models
type Room struct {
	ID        int
	RoomName  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Restriction is the restrictions models
type Restriction struct {
	ID              int
	RestrictionName string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Reservations is the reservation model
type Reservation struct {
	ID        int
	FirstName string
	LastName  string
	Email     string
	Phone     string
	StartDate time.Time
	EndDate   time.Time
	RoomID    int
	CreatedAt time.Time
	UpdatedAt time.Time
	Room      Room
}

// Restrictions is the rooms restriction mod3l
type RoomRestriction struct {
	ID            int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	StartDate     time.Time
	EndDate       time.Time
	RoomID        int
	ReservationID int
	RestrictionID int
	Room          Room
	Reservation   Reservation
	Restriction   Restriction
}

// GoogleUser represents user information obtained from Google OAuth
type GoogleUser struct {
	ID           int
	FirstName    string
	LastName     string
	Email        string
	AccessToken  string // Google OAuth access token
	RefreshToken string // Google OAuth refresh token (if needed)
	ExpiresAt    time.Time // Access token expiration time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}







