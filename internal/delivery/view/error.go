package view

import (
	"fmt"
	"net/http"

	"github.com/taosx/interview-challenge/internal/templatemanager"
)

type ViewError struct {
	ErrorMsg   string
	ErrorCause error
}

func (ve ViewError) Error() string {
	return fmt.Sprintf("%s: %s", ve.ErrorMsg, ve.ErrorCause.Error())
}

func (ve ViewError) handler(w http.ResponseWriter, r *http.Request) {
	err := templatemanager.RenderTemplate(w, "error.html", ve)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Sever: template couldn't be rendered" + err.Error()))
	}
}
