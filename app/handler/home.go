package handler

import (
	"net/http"

	"github.com/pmmp/CrashArchive/app/view"
)

func HomeGet(w http.ResponseWriter, r *http.Request) {
	view.ExecuteTemplate(w, "home", nil)
}
