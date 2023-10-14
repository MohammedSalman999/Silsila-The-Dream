package dbrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/mohammedsalman999/silsila/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (m *mysqlDBRepo) ALLUsers() bool {
	return true

}

// InsertRservation implements repository.DatabaseRepo.
func (m *mysqlDBRepo) InsertReservation(res models.Reservation) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var newID int
	stmt := `
		INSERT INTO reservations (first_name, last_name, email, phone, start_date, end_date, room_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Executing the INSERT statement
	result, err := m.DB.ExecContext(ctx, stmt,
		res.FirstName,
		res.LastName,
		res.Email,
		res.Phone,
		res.StartDate,
		res.EndDate,
		res.RoomID,
		time.Now(),
		time.Now(),
	)

	if err != nil {
		return 0, err
	}

	// Retrieve the last insert ID
	newID64, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	newID = int(newID64) // Convert int64 to int

	return newID, nil
}

// Insert A Room Restriction Into The DataBase
func (m *mysqlDBRepo) InsertRoomRestrictions(r models.RoomRestriction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
        INSERT INTO room_restrictions (start_date, end_date, room_id, reservation_id, created_at, updated_at, restriction_id)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `

	_, err := m.DB.ExecContext(ctx, stmt,
		r.StartDate,
		r.EndDate,
		r.RoomID,
		r.ReservationID,
		time.Now(),
		time.Now(),
		r.RestrictionID,
	)

	if err != nil {
		return err
	}
	return nil
}

// SearchAvailability returns true if availabale else thinga dikha deti hai
func (m *mysqlDBRepo) SearchAvailabilityByRoomID(start, end time.Time, roomID int) (bool, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    query := `
        SELECT 
            COUNT(id)
        FROM 
            room_restrictions
        WHERE
            room_id = ? 
            AND ? <= end_date AND ? >= start_date ;
    `

    var numRows int
    row := m.DB.QueryRowContext(ctx, query, roomID, start, end)
    err := row.Scan(&numRows)
    if err != nil {
        if err == sql.ErrNoRows {
            // No overlapping reservations found, room is available
            return true, nil
        }
        return false, err
    }

    if numRows == 0 {
        return true, nil
    }
    return false, nil
}


// SearchRoomAvailabilityForAllRooms returns a slice of available rooms for a given date range
func (m *mysqlDBRepo) SearchRoomAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var rooms []models.Room

    query := `
        SELECT
            r.id,
            r.room_name
        FROM
            rooms r
        WHERE
            r.id NOT IN (
                SELECT room_id FROM room_restrictions rr WHERE ? < rr.end_date AND ? > rr.start_date
            );
    `

    m.App.InfoLog.Printf("Executing query with start: %s and end: %s\n", start, end)

    rows, err := m.DB.QueryContext(ctx, query, start, end)
    if err != nil {
        return rooms, err
    }

    for rows.Next() {
        var room models.Room
        err := rows.Scan(
            &room.ID,
            &room.RoomName,
        )
        if err != nil {
            m.App.ErrorLog.Println("Error scanning row:", err)
            return rooms, err
        }
        m.App.InfoLog.Println("Fetched Room:", room.ID, room.RoomName)
        rooms = append(rooms, room)
    }

    if err = rows.Err(); err != nil {
        return rooms, err
    }

    return rooms, nil
}

//GetRoomByID gets a room by id 
func (m *mysqlDBRepo) GetRoomByID(id int) (models.Room, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var room models.Room

    query := `SELECT id, room_name, created_at, updated_at FROM rooms WHERE id = ?`

    row := m.DB.QueryRowContext(ctx, query, id)

    err := row.Scan(
        &room.ID,
        &room.RoomName,
        &room.CreatedAt,
        &room.UpdatedAt,
    )
    if err != nil {
        if err == sql.ErrNoRows {
            return room, fmt.Errorf("room not found")
        }
        return room, err
    }

    return room, nil
}

//Gets the user by userid
func ( m *mysqlDBRepo) GetUserById(id int) (models.User, error){
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    query := `select id , first_name , last_name , email , password , access_level , created_at,
        uopdate_at from users where id = ?`


    row := m.DB.QueryRowContext(ctx,query,id)

    var u models.User
    err := row.Scan(
        &u.ID,
        &u.FirstName,
        &u.LastName,
        &u.Email,
        &u.Password,
        &u.AccessLevel,
        &u.CreatedAt,
        &u.UpdatedAt,
    )
    if err!= nil{
        return u,err
    }
    return u ,nil
}

// update user updates the user 

func (m *mysqlDBRepo) UpdateUser ( u models.User) error {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    query := `
    updateusers set first_name = ? , last_naem = ? , email = ? , access_level = ? ,updated_at =?
    
    
    
    `

    _,err := m.DB.ExecContext(ctx,query,
        u.FirstName,
        u.LastName,
        u.Email,
        u.AccessLevel,
        time.Now(),    
        
    )

    if err != nil{
        return err
    }
    return nil
   
}


func (m *mysqlDBRepo) Authenticate(email, testPassword string) (int, string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var id int
    var hashedPassword string

    row := m.DB.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email=?", email)
    err := row.Scan(&id, &hashedPassword)

    if err != nil {
        return id, "", err
    }

    err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))

    if err == bcrypt.ErrMismatchedHashAndPassword {
        return 0, "", errors.New("incorrect password")
    } else if err != nil {
        return 0, "", err
    }

    return id, hashedPassword, nil
}

// Inserts the user with hashed password
func (m *mysqlDBRepo) InsertUser(user models.User) (int, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var newID int
    stmt := `
        INSERT INTO users (first_name, last_name, email, password, access_level, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?, ?)
    `

    // Hash the password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return 0, err
    }

    // Executing the INSERT statement
    result, err := m.DB.ExecContext(ctx, stmt,
        user.FirstName,
        user.LastName,
        user.Email,
        hashedPassword, // Store the hashed password
        user.AccessLevel,
        time.Now(),
        time.Now(),
    )

    if err != nil {
        return 0, err
    }

    // Retrieve the last insert ID
    newID64, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }
    newID = int(newID64) // Convert int64 to int

    return newID, nil
}


// VerifyLogin verifies if the entered email and password match a user in the database.
func (m *mysqlDBRepo) VerifyLogin(email, password string) (int, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
    defer cancel()

    var userID int
    var hashedPassword string

    // Retrieve user details by email
    row := m.DB.QueryRowContext(ctx, "SELECT id, password FROM users WHERE email = ?", email)
    err := row.Scan(&userID, &hashedPassword)
    if err != nil {
        return 0, err
    }

    // Compare stored hashed password with entered password
    err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    if err != nil {
        return 0, err
    }

    return userID, nil
}














