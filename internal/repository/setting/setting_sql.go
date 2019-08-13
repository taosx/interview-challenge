package setting

import (
	"log"

	"github.com/taosx/interview-challenge/internal/domain"
	"github.com/taosx/interview-challenge/internal/storage"
)

type repoSQLite struct {
	db *storage.SQLiteStorage
}

func NewSQLiteRepo(db *storage.SQLiteStorage) SettingRepository {
	repo := &repoSQLite{
		db: db,
	}

	return repo.autoMigrate().initialPopulate()
}

func (r repoSQLite) GetTitle() (string, error) {
	var title string

	err := r.db.Get(&title, "SELECT value FROM setting WHERE name='title' LIMIT 1;")
	if err != nil {
		return "", err
	}

	return title, nil
}

func (r repoSQLite) GetNavigation() ([]domain.Link, error) {
	ls := links{}

	err := r.db.Get(&ls, "SELECT value FROM setting WHERE name='navigation' LIMIT 1;")
	if err != nil {
		return nil, err
	}

	return ls.toDomain(), nil
}

func (r repoSQLite) addNavigationLink(linkURL, title string) error {
	ls := links{}
	err := r.db.Get(&ls, "SELECT value FROM setting WHERE name='navigation' LIMIT 1;")
	if err != nil {
		return err
	}

	l := link{
		Title: title,
		URL:   linkURL,
	}

	ls = append(ls, l)

	_, err = r.db.Exec(`
	UPDATE setting SET
		value = $1
	WHERE name = 'navigation';`, ls)
	if err != nil {
		return err
	}

	return nil
}

func (r *repoSQLite) autoMigrate() *repoSQLite {
	q := `
	CREATE TABLE IF NOT EXISTS
	setting
	(
		name VARCHAR(128) PRIMARY KEY UNIQUE,
		value VARCHAR(512) DEFAULT NULL
	);
	`

	_, err := r.db.Exec(q)
	if err != nil {
		log.Fatalln("migration failed: " + err.Error())
	}

	return r
}

func (r *repoSQLite) initialPopulate() *repoSQLite {
	q := `
	INSERT OR REPLACE INTO setting (
		name,
		value
	) VALUES (
		'title',
		'Ticketing System'
	);

	INSERT OR REPLACE INTO setting (
		name,
		value
	) VALUES (
		'navigation',
		''
	);
	`
	_, err := r.db.Exec(q)
	if err != nil {
		log.Fatalln("settings population failed: " + err.Error())
	}

	err = r.addNavigationLink("/", "Home")
	if err != nil {
		log.Fatalln("navigation settings population failed: " + err.Error())
	}

	err = r.addNavigationLink("/guestlist", "Guests")
	if err != nil {
		log.Fatalln("navigation settings population failed: " + err.Error())
	}

	return r
}
