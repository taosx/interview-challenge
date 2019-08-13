package engine

import (
	"github.com/taosx/interview-challenge/internal/payment/processorStripe"

	"github.com/taosx/interview-challenge/internal/domain"
)

type ticketRepo interface {
	Reserve(guestID int) (*domain.Ticket, error)
	AttachCheckoutSessionID(ticketID int, sessionID string) error
	IsAlreadyReservedErr(err error) bool
	GetReservedBySessionID(sessionID string) (*domain.Ticket, error)
	Book(ticketID int) error
}

type userRepository interface {
	Create(name string) (*domain.User, error)
	GetByName(name string) (*domain.User, error)
	IsDuplicateErr(err error) bool
}

type paymentProcessor interface {
	CreateCheckoutSession(successPath, cancelPath string) *processorStripe.Session
}

type Environment struct {
	TicketRepo       ticketRepo
	UserRepo         userRepository
	PaymentProcessor paymentProcessor
}

type Engine struct{ Environment }

func (e Environment) New() *Engine {
	return &Engine{e}
}
