package view

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/taosx/interview-challenge/internal/domain"

	"github.com/go-chi/chi"

	"github.com/taosx/interview-challenge/internal/templatemanager"
)

type BookingPageData struct {
	MainData
	TicketID          int
	TicketPrice       int
	UserName          string
	CheckoutSessionID string
}

func (v *Viewer) BookingHandler(w http.ResponseWriter, r *http.Request) {
	nameSlug := chi.URLParam(r, "userSlug")
	user, userErr := v.userRepo.GetBySlug(nameSlug)
	if userErr != nil {
		ViewError{ErrorCause: userErr, ErrorMsg: "couldn't retrieve user with name slug"}.handler(w, r)
		return
	}

	var ticket *domain.Ticket
	var err error

	ticker := time.NewTicker(time.Millisecond * 500)
	timer := time.NewTimer(time.Second * 10)

	for {
		select {
		case <-ticker.C:
			ticket, err = v.ticketRepo.GetReservedByUserID(user.ID)
			if err != nil {
				ViewError{ErrorCause: err, ErrorMsg: "couldn't retrieve reserved ticket by user"}.handler(w, r)
				return
			}

			if ticket.State == domain.Booked {
				if !timer.Stop() {
					<-timer.C
				}
				ticker.Stop()
				break
			}
		case <-timer.C:
			ViewError{ErrorCause: errors.New("Payment couldn't be confirmed"), ErrorMsg: "Couldn't book this ticket"}.handler(w, r)
			return
		}
		break
	}

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

	data := BookingPageData{
		MainData: MainData{
			WebTitle:        title,
			NavigationLinks: navigationLinks,
		},
		TicketID:          ticket.ID,
		TicketPrice:       ticket.Cost,
		UserName:          user.Name,
		CheckoutSessionID: ticket.SessionID,
	}

	err = templatemanager.RenderTemplate(w, "page-booking.html", data)
	if err != nil {
		ViewError{ErrorCause: err, ErrorMsg: "couldn't render 'page-booking' template"}.handler(w, r)
		log.Println(err)
	}
}
