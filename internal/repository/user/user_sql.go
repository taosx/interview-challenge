package user

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/gosimple/slug"

	"github.com/taosx/interview-challenge/internal/domain"
	"github.com/taosx/interview-challenge/internal/storage"
)

type repoSQLite struct {
	db *storage.SQLiteStorage
}

func NewSQLiteRepo(db *storage.SQLiteStorage) UserRepository {
	repo := &repoSQLite{
		db: db,
	}

	return repo.autoMigrate()
}

func (r repoSQLite) Create(name string) (*domain.User, error) {
	nameSlug := slug.Make(name)

	result, err := r.db.Exec("INSERT INTO user (name, name_slug) VALUES ($1, $2);", name, nameSlug)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, errors.New("guest '" + name + "' couldn't be created, reason unknown")
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	user := (&userDB{
		ID:       int(userID),
		Name:     name,
		NameSlug: nameSlug,
	}).toDomain()

	return &user, nil
}

func (r repoSQLite) GetGuests() ([]domain.User, error) {
	users := usersDB{}

	err := r.db.Select(&users, "SELECT u.* FROM user AS u INNER JOIN ticket AS t ON t.user_id = u.id;")
	if err != nil {
		return nil, err
	}

	return users.toDomain(), err
}

func (r repoSQLite) GetByID(userID int) (*domain.User, error) {
	user := new(userDB)
	err := r.db.Get(user, "SELECT * FROM user WHERE id = $1;", userID)
	if err != nil {
		return nil, err
	}

	domainUser := user.toDomain()
	return &domainUser, nil
}

func (r repoSQLite) GetByName(name string) (*domain.User, error) {
	nameSlug := slug.Make(name)

	user := new(userDB)
	err := r.db.Get(user, "SELECT * FROM user WHERE name_slug = $1;", nameSlug)
	if err != nil {
		return nil, err
	}

	domainUser := user.toDomain()
	return &domainUser, nil
}

func (r repoSQLite) GetBySlug(nameSlug string) (*domain.User, error) {
	user := new(userDB)
	err := r.db.Get(user, "SELECT * FROM user WHERE name_slug = $1;", nameSlug)
	if err != nil {
		return nil, err
	}

	domainUser := user.toDomain()
	return &domainUser, nil
}

func (r repoSQLite) IsDuplicateErr(err error) bool {
	if err.Error() == "UNIQUE constraint failed: user.name_slug" {
		return true
	}
	return false
}

func getByID(db sqlx.Ext, userID int) (*userDB, error) {
	user := new(userDB)
	err := db.QueryRowx("SELECT * FROM user WHERE id = $1;", userID).StructScan(user)
	if err != nil {
		return nil, err
	}
	return user, err
}

func (r repoSQLite) autoMigrate() repoSQLite {
	q := `
	CREATE TABLE IF NOT EXISTS
	user
	(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		name_slug TEXT UNIQUE
	);
	CREATE UNIQUE INDEX IF NOT EXISTS uniqueNameSlug ON user(name_slug);
	`

	_, err := r.db.Exec(q)
	if err != nil {
		log.Fatalln("migration failed: " + err.Error())
	}

	return r
}
