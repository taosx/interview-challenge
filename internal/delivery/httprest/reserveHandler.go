package httprest

import (
	"encoding/json"
	"net/http"
	"regexp"
	"unicode"
)

type reserveRequestBody struct {
	Name string `json:"name"`
}

type reserveResponseBody struct {
	UserSlug string `json:"user_slug"`
}

var isValid = regexp.MustCompile(`(?m)^([A-Za-z\\u00D8-\\u00f6\\u00f8-\\u00ff\s]*)$`).MatchString

// TODO: return json
func (h Handlers) ReserveHandler(w http.ResponseWriter, r *http.Request) {
	reqBody := new(reserveRequestBody)
	err := json.NewDecoder(r.Body).Decode(reqBody)
	if err != nil {
		return
	}

	if len(reqBody.Name) >= 70 {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("name too long, the maximum length a name can have is 70 characters"))
		return
	}

	if !isValid(reqBody.Name) {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("name can contain only letters and whitespace"))
		return
	}

	nameSlug, err := h.e.MakeReservation(reqBody.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	if nameSlug == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to make reservation"))
		return
	}

	w.WriteHeader(http.StatusOK)
	respBody := reserveResponseBody{
		UserSlug: nameSlug,
	}

	if err := json.NewEncoder(w).Encode(&respBody); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to encode success json"))
		return
	}

	// msg := fmt.Sprintf("Reservation for ticket: '%d' complete", ticketID)
}

func isLatin(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			if !unicode.IsSpace(r) {
				return false
			}
		}
	}
	return true
}
