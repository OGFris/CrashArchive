package handler

import (
	"errors"
	"log"
	"net/http"

	"strconv"
)

const pageSize = 50

type List struct{ *Common }

func (l List) Get(w http.ResponseWriter, r *http.Request) {
	pageID, err := PageID(r.URL.Query().Get("page"))
	if err != nil {
		log.Println(err)
		l.View.Error(w, "", http.StatusNotFound)
		return
	}

	total, start, reports, err := l.DB.GetFilteredReports(pageID, pageSize, "WHERE duplicate = false")
	if err != nil {
		log.Println(err)
		l.View.Error(w, "", http.StatusInternalServerError)
		return
	}

	l.View.ExecuteList(w, reports, r.URL.String(), pageID, start, total)
}

func PageID(page string) (int, error) {
	var pageID int = 1
	var err error

	if page != "" {
		pageID, err = strconv.Atoi(page)
		if err != nil || pageID <= 0 {
			return pageID, errors.New("invalid pageID")
		}
	}
	return pageID, nil
}
