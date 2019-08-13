package user

import (
	"github.com/taosx/interview-challenge/internal/domain"
)

type UserRepository interface {
	Create(name string) (*domain.User, error)
	GetByID(userID int) (*domain.User, error)
	GetByName(name string) (*domain.User, error)
	GetBySlug(nameSlug string) (*domain.User, error)
	GetGuests() ([]domain.User, error)
	IsDuplicateErr(err error) bool
}

type userDB struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	NameSlug string `db:"name_slug"`
}
type usersDB []userDB

func (user *userDB) toDomain() domain.User {
	return domain.User{
		ID:       user.ID,
		Name:     user.Name,
		NameSlug: user.NameSlug,
	}
}

func (users usersDB) toDomain() []domain.User {
	domainUsers := make([]domain.User, len(users), len(users))

	for inx, _ := range users {
		domainUsers[inx] = domain.User{
			ID:       users[inx].ID,
			Name:     users[inx].Name,
			NameSlug: users[inx].NameSlug,
		}
	}

	return domainUsers
}
