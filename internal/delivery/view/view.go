package view

import (
	"github.com/taosx/interview-challenge/internal/domain"
)

type settingsRepository interface {
	GetTitle() (string, error)
	GetNavigation() ([]domain.Link, error)
}

type ticketRepository interface {
	GetReservedByID(id int) (*domain.Ticket, error)
	GetReservedByUserID(userID int) (*domain.Ticket, error)
	CountUnreserved() (int, error)
}

type userRepository interface {
	GetGuests() ([]domain.User, error)
	GetByID(id int) (*domain.User, error)
	GetBySlug(nameSlug string) (*domain.User, error)
}

type Viewer struct {
	configRepo settingsRepository
	ticketRepo ticketRepository
	userRepo   userRepository
}

func NewViewLayer(
	configRepo settingsRepository,
	ticketRepo ticketRepository,
	userRepo userRepository,
) *Viewer {
	return &Viewer{
		configRepo: configRepo,
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
}
