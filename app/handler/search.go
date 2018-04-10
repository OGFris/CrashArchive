package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/pmmp/CrashArchive/app/crashreport"
)

type Search struct{ *Common }

func (s Search) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", s.Get)
	r.Get("/id", s.ID)
	r.Get("/plugin", s.Plugin)
	r.Get("/build", s.Build)
	r.Get("/report", s.Report)
	return r
}
func (s Search) Get(w http.ResponseWriter, r *http.Request) {
	s.View.Execute(w, "search", nil)
}

func (s Search) ID(w http.ResponseWriter, r *http.Request) {
	reportID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/view/%d", reportID), http.StatusMovedPermanently)
}

func (s Search) Plugin(w http.ResponseWriter, r *http.Request) {
	plugin := r.URL.Query().Get("plugin")
	if plugin == "" {
		log.Println("empty plugin name")
		s.View.Error(w, "", http.StatusBadRequest)
		return
	}

	pageID, err := PageID(r.URL.Query().Get("page"))
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusNotFound)
		return
	}

	total, start, reports, err := s.DB.GetFilteredReports(pageID, pageSize, "WHERE plugin = ?", plugin)
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusInternalServerError)
		return
	}

	s.View.ExecuteList(w, reports, r.URL.String(), pageID, start, total)
}

func (s Search) Build(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	buildID, err := strconv.Atoi(params.Get("build"))
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusBadRequest)
		return
	}

	var operator string

	switch params.Get("type") {
	case "gt":
		operator = ">"
	case "lt":
		operator = "<"
	default:
		operator = "="
	}

	pageID, err := PageID(r.URL.Query().Get("page"))
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusNotFound)
		return
	}

	total, start, reports, err := s.DB.GetFilteredReports(pageID, pageSize, fmt.Sprintf("WHERE build %s ?", operator), buildID)
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusInternalServerError)
		return
	}

	s.View.ExecuteList(w, reports, r.URL.String(), pageID, start, total)
}

func (s Search) Report(w http.ResponseWriter, r *http.Request) {
	query := "SELECT * FROM crash_reports WHERE id = ?"
	reportID, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusBadRequest)
		return
	}

	var report crashreport.Report
	err = s.DB.Get(&report, query, reportID)
	if err != nil {
		log.Println(err)
		s.View.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	pageID, err := PageID(r.URL.Query().Get("page"))
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusNotFound)
		return
	}

	total, start, reports, err := s.DB.GetFilteredReports(pageID, pageSize, "WHERE message = ? AND file = ? and line = ?", report.Message, report.File, report.Line)
	if err != nil {
		log.Println(err)
		s.View.Error(w, "", http.StatusInternalServerError)
		return
	}

	s.View.ExecuteList(w, reports, r.URL.String(), pageID, start, total)
}
