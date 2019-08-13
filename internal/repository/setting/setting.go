package setting

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"errors"

	"github.com/taosx/interview-challenge/internal/domain"
)

type SettingRepository interface {
	GetTitle() (string, error)
	GetNavigation() ([]domain.Link, error)
}

type link struct {
	URL   string `db:"url"`
	Title string `db:"title"`
}

type links []link

func (l link) toDomain() domain.Link {
	return domain.Link{
		Title: l.Title,
		URL:   l.URL,
	}
}

func (ls links) toDomain() []domain.Link {
	domainLinks := make([]domain.Link, len(ls), len(ls))

	for inx, _ := range ls {
		domainLinks[inx] = domain.Link{
			Title: ls[inx].Title,
			URL:   ls[inx].URL,
		}
	}

	return domainLinks
}

func (ls links) Value() (driver.Value, error) {
	buf := new(bytes.Buffer)

	err := gob.NewEncoder(buf).Encode(&ls)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (ls *links) Scan(value interface{}) error {
	if value == nil {
		// set the value of the pointer yne to YesNoEnum(false)
		*ls = nil
		return nil
	}

	if value == "" {
		*ls = nil
		return nil
	}

	var reader *bytes.Reader
	switch value.(type) {
	case string:
		reader = bytes.NewReader([]byte(value.(string)))
	case []byte:
		reader = bytes.NewReader(value.([]byte))
	default:
		return errors.New("couldn't decode type different that string or bytes")
	}

	err := gob.NewDecoder(reader).Decode(ls)
	if err != nil {
		return err
	}

	return nil
}
