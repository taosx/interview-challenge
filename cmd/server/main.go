package main

import (
	"log"
	"net/http"

	"github.com/taosx/interview-challenge/internal/payment/processorStripe"

	"github.com/taosx/interview-challenge/internal/repository/setting"
	"github.com/taosx/interview-challenge/internal/repository/user"

	"github.com/go-chi/chi"

	"github.com/taosx/interview-challenge/internal/delivery/httprest"
	"github.com/taosx/interview-challenge/internal/delivery/view"
	"github.com/taosx/interview-challenge/internal/engine"
	"github.com/taosx/interview-challenge/internal/repository/ticket"
	"github.com/taosx/interview-challenge/internal/storage"
	"github.com/taosx/interview-challenge/internal/templatemanager"
)

func main() {

	templatemanager.SetTemplateConfig("web/template/layouts/", "web/template/")
	err := templatemanager.LoadTemplates()
	if err != nil {
		log.Fatalln(err)
	}

	sqliteDB := storage.NewStorageSQLite("__deleteme.db")

	ticketRepo := ticket.NewSQLiteRepo(sqliteDB)
	settingRepo := setting.NewSQLiteRepo(sqliteDB)
	userRepo := user.NewSQLiteRepo(sqliteDB)

	processor := processorStripe.New(
		"http://8bac8c99.ngrok.io",
		"sk_test_88SK2xCVnzLYO031KfX9jYy400VK2yuEdj",
	)

	engine := engine.Environment{
		TicketRepo:       ticketRepo,
		UserRepo:         userRepo,
		PaymentProcessor: processor,
	}.New()

	viewLayer := view.NewViewLayer(
		settingRepo,
		ticketRepo,
		userRepo,
	)

	rest := httprest.NewRestLayer(engine)

	r := chi.NewRouter()
	r.Get("/", viewLayer.IndexHandler)
	r.Get("/guestlist", viewLayer.UsersHandler)
	r.Get("/booking/{userSlug}", viewLayer.BookingHandler)
	r.Get("/booked", BookPage)
	r.Post("/api/reserve", rest.ReserveHandler)
	r.Post("/api/webhook/stripe", rest.WebHookStripe)

	log.Fatalln(http.ListenAndServe("localhost:3000", r))
}

func BookPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Booked Page"))
	return
}
