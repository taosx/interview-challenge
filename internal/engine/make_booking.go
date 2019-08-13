package engine

import (
	"github.com/pkg/errors"
)

func (e *Engine) MakeBooking(sessionID string) error {
	if sessionID == "" {
		return errors.New("sessionID is needed in order to book reservation")
	}

	ticket, err := e.TicketRepo.GetReservedBySessionID(sessionID)
	if err != nil {
		return errors.Wrapf(err, "failed to book ticket with sessionID '%s'", sessionID)
	}

	err = e.TicketRepo.Book(ticket.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to book ticket '%d' with sessionID '%s'", ticket.ID, sessionID)
	}

	return nil
}
