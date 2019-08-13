package view

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/taosx/interview-challenge/internal/domain"

	"github.com/taosx/interview-challenge/internal/templatemanager"
)

type Guest struct {
	Name string
}

type GuestPageData struct {
	MainData
	Guests []domain.User
}

func (v *Viewer) UsersHandler(w http.ResponseWriter, r *http.Request) {
	title, err := v.configRepo.GetTitle()
	if err != nil {
		ViewError{ErrorCause: err, ErrorMsg: "couldn't get title from db"}.handler(w, r)
		return
	}

	navigationLinks, err := v.configRepo.GetNavigation()
	if err != nil && err != sql.ErrNoRows {
		ViewError{ErrorCause: err, ErrorMsg: "couldn't get pages from db"}.handler(w, r)
		return
	}

	guestList, err := v.userRepo.GetGuests()
	if err != nil && err != sql.ErrNoRows {
		ViewError{ErrorCause: err, ErrorMsg: "couldn't get guests from db"}.handler(w, r)
		return
	}

	data := GuestPageData{
		MainData: MainData{
			WebTitle:        title,
			NavigationLinks: navigationLinks,
		},
		Guests: guestList,
	}

	err = templatemanager.RenderTemplate(w, "page-guestlist.html", data)
	if err != nil {
		ViewError{ErrorCause: err, ErrorMsg: "couldn't render 'page-guestlist' template"}.handler(w, r)
		log.Println(err)
	}
}
