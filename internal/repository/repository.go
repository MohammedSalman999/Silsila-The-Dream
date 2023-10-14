package repository

import (
	"time"

	"github.com/mohammedsalman999/silsila/internal/models"
)

type DatabaseRepo interface {
	ALLUsers() bool

	InsertReservation(res models.Reservation) (int,error)
	InsertRoomRestrictions(r models.RoomRestriction) error
	SearchAvailabilityByRoomID(start, end time.Time, roomID int) (bool, error)
	SearchRoomAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error)
	GetRoomByID(id int) (models.Room ,error)
	GetUserById(id int) (models.User, error)
	Authenticate(email,testPassword string ) (int,string,error)
	InsertUser(user models.User) (int, error)
	VerifyLogin(email, password string) (int, error)
}