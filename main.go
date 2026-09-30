// Command surf-api is a small surf lesson booking service built with GoFr.
package main

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"gofr.dev/pkg/gofr"
	gofrHTTP "gofr.dev/pkg/gofr/http"

	"github.com/ajassaif/surf-api/migrations"
)

// Slots a lesson can be booked in. Varkala surf is best early morning and late afternoon.
var validSlots = map[string]bool{"morning": true, "evening": true}

type Instructor struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
}

type Booking struct {
	ID           int64  `json:"id"`
	GuestName    string `json:"guest_name"`
	InstructorID int    `json:"instructor_id"`
	Date         string `json:"date"`
	Slot         string `json:"slot"`
}

func main() {
	app := gofr.New()

	app.Migrate(migrations.All())

	app.GET("/instructors", listInstructors)
	app.GET("/bookings", listBookings)
	app.POST("/bookings", createBooking)
	app.DELETE("/bookings/{id}", cancelBooking)

	app.Run()
}

func listInstructors(c *gofr.Context) (any, error) {
	rows, err := c.SQL.QueryContext(c, "SELECT id, name, specialty FROM instructors ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instructors := []Instructor{}

	for rows.Next() {
		var i Instructor
		if err := rows.Scan(&i.ID, &i.Name, &i.Specialty); err != nil {
			return nil, err
		}

		instructors = append(instructors, i)
	}

	return instructors, rows.Err()
}

// listBookings returns bookings, optionally filtered by ?date=YYYY-MM-DD.
func listBookings(c *gofr.Context) (any, error) {
	query := "SELECT id, guest_name, instructor_id, lesson_date, slot FROM bookings"
	args := []any{}

	if date := c.Param("date"); date != "" {
		if _, err := time.Parse(time.DateOnly, date); err != nil {
			return nil, gofrHTTP.ErrorInvalidParam{Params: []string{"date"}}
		}

		query += " WHERE lesson_date = ?"
		args = append(args, date)
	}

	rows, err := c.SQL.QueryContext(c, query+" ORDER BY lesson_date, slot", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := []Booking{}

	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.GuestName, &b.InstructorID, &b.Date, &b.Slot); err != nil {
			return nil, err
		}

		bookings = append(bookings, b)
	}

	return bookings, rows.Err()
}

func createBooking(c *gofr.Context) (any, error) {
	var b Booking
	if err := c.Bind(&b); err != nil {
		return nil, gofrHTTP.ErrorInvalidParam{Params: []string{"body"}}
	}

	b.GuestName = strings.TrimSpace(b.GuestName)
	b.Slot = strings.ToLower(strings.TrimSpace(b.Slot))

	var missing []string

	if b.GuestName == "" {
		missing = append(missing, "guest_name")
	}

	if b.InstructorID == 0 {
		missing = append(missing, "instructor_id")
	}

	if b.Date == "" {
		missing = append(missing, "date")
	}

	if b.Slot == "" {
		missing = append(missing, "slot")
	}

	if len(missing) > 0 {
		return nil, gofrHTTP.ErrorMissingParam{Params: missing}
	}

	if _, err := time.Parse(time.DateOnly, b.Date); err != nil {
		return nil, gofrHTTP.ErrorInvalidParam{Params: []string{"date"}}
	}

	if !validSlots[b.Slot] {
		return nil, gofrHTTP.ErrorInvalidParam{Params: []string{"slot"}}
	}

	// Make sure the instructor exists before booking them.
	var exists int

	err := c.SQL.QueryRowContext(c, "SELECT 1 FROM instructors WHERE id = ?", b.InstructorID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, gofrHTTP.ErrorEntityNotFound{Name: "instructor_id", Value: strconv.Itoa(b.InstructorID)}
	}

	if err != nil {
		return nil, err
	}

	res, err := c.SQL.ExecContext(c,
		"INSERT INTO bookings (guest_name, instructor_id, lesson_date, slot) VALUES (?, ?, ?, ?)",
		b.GuestName, b.InstructorID, b.Date, b.Slot)
	if err != nil {
		// The UNIQUE constraint stops double-booking an instructor.
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, gofrHTTP.ErrorEntityAlreadyExist{}
		}

		return nil, err
	}

	b.ID, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return b, nil
}

func cancelBooking(c *gofr.Context) (any, error) {
	id := c.PathParam("id")
	if _, err := strconv.Atoi(id); err != nil {
		return nil, gofrHTTP.ErrorInvalidParam{Params: []string{"id"}}
	}

	res, err := c.SQL.ExecContext(c, "DELETE FROM bookings WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	if n, _ := res.RowsAffected(); n == 0 {
		return nil, gofrHTTP.ErrorEntityNotFound{Name: "id", Value: id}
	}

	return nil, nil
}
