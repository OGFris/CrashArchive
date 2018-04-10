package handler

import (
	"net/http"
)

type Home struct{ *Common }

func (h Home) Get(w http.ResponseWriter, r *http.Request) {
	h.View.Execute(w, "home", nil)
}
