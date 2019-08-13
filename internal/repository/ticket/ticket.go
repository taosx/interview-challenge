package ticket

import (
	"database/sql"
	"time"

	"github.com/taosx/interview-challenge/internal/domain"
)

type TicketRepository interface {
	CountUnreserved() (int, error)
	Reserve(guestID int) (*domain.Ticket, error)
	Book(ticketID int) error
	AttachCheckoutSessionID(ticketID int, sessionID string) error
	IsAlreadyReservedErr(err error) bool
	GetReservedByID(ticketID int) (*domain.Ticket, error)
	GetReservedByUserID(userID int) (*domain.Ticket, error)
	GetReservedBySessionID(sessionID string) (*domain.Ticket, error)
}

type ticket_state string

const (
	Unreserved ticket_state = "unreserved"
	Reserved   ticket_state = "reserved"
	Booked     ticket_state = "booked"
)

type Ticket struct {
	ID        int            `db:"id"`
	Cost      int            `db:"cost"`
	State     ticket_state   `db:"state"`
	UserID    sql.NullInt64  `db:"user_id"`
	SessionID sql.NullString `db:"session_id"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

func (t *Ticket) toDomain() domain.Ticket {
	var sessionID string
	if t.SessionID.Valid {
		valueSessionID, err := t.SessionID.Value()
		if err != nil {
			sessionID = ""
		} else {
			sessionID = valueSessionID.(string)
		}
	}

	domainTicket := domain.Ticket{
		ID:        t.ID,
		Cost:      t.Cost,
		UserID:    t.UserID,
		SessionID: sessionID,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}

	switch t.State {
	case Unreserved:
		domainTicket.State = domain.Unreserved
	case Reserved:
		domainTicket.State = domain.Reserved
	case Booked:
		domainTicket.State = domain.Booked
	}

	return domainTicket
}
