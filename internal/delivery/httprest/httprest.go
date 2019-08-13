package httprest

type Engine interface {
	MakeReservation(guestName string) (nameSlug string, err error)
	MakeBooking(sessionID string) error
}

type Handlers struct {
	e Engine
}

func NewRestLayer(e Engine) *Handlers {
	return &Handlers{
		e: e,
	}
}
