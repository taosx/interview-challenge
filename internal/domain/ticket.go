package domain

import (
	"database/sql"
	"time"
)

type ticket_state string

const (
	Unreserved ticket_state = "unreserved"
	Reserved   ticket_state = "reserved"
	Booked     ticket_state = "booked"
)

type Ticket struct {
	ID        int
	Cost      int
	State     ticket_state
	UserID    sql.NullInt64
	SessionID string
	CreatedAt time.Time
	UpdatedAt time.Time
}
