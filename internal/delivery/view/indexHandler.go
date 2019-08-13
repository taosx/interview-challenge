package view

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/taosx/interview-challenge/internal/domain"
	"github.com/taosx/interview-challenge/internal/templatemanager"
)

type MainData struct {
	WebTitle        string
	TicketCount     int
	NavigationLinks []domain.Link
}

func (v *Viewer) IndexHandler(w http.ResponseWriter, r *http.Request) {
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

	ticketsLeft, err := v.ticketRepo.CountUnreserved()
	if err != nil {
		ViewError{ErrorCause: err, ErrorMsg: "couldn't get pages from db"}.handler(w, r)
		return
	}

	data := MainData{
		WebTitle:        title,
		NavigationLinks: navigationLinks,
		TicketCount:     ticketsLeft,
	}

	err = templatemanager.RenderTemplate(w, "page-home.html", data)
	if err != nil {
		ViewError{ErrorCause: err, ErrorMsg: "couldn't render 'page-home' template"}.handler(w, r)
		log.Println(err)
	}
}
