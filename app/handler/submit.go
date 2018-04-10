package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/pmmp/CrashArchive/app/crashreport"
)

type Submit struct{ *Common }

func (s Submit) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", s.Get)
	r.Post("/", s.Post)
	r.Post("/api", s.Post)
	return r
}

func (s Submit) Get(w http.ResponseWriter, r *http.Request) {
	s.View.Execute(w, "submit", nil)
}

func multipartForm(r *http.Request) (string, error) {
	if r.FormValue("report") != "yes" {
		return "", errors.New("report is not yes")
	}

	if err := r.ParseMultipartForm(1024 * 256); err != nil {
		return "", err
	}

	return ParseMultipartForm(r.MultipartForm)
}

func (s Submit) Post(w http.ResponseWriter, r *http.Request) {
	reportStr, err := multipartForm(r)
	if err != nil {
		http.Redirect(w, r, "/submit", http.StatusMovedPermanently)
	}

	isAPI := strings.HasSuffix(r.RequestURI, "/api")

	report, err := crashreport.DecodeCrashReport(reportStr)
	if err != nil {
		log.Printf("got invalid crash report from: %s (%v)", r.RemoteAddr, err)
		s.sendError(w, "This crash report is not valid", http.StatusUnprocessableEntity, isAPI)
		return
	}

	if err := report.Data.IsValid(); err != nil {
		log.Printf("%s from: %s (%v)", err, r.RemoteAddr)
		s.sendError(w, "", http.StatusTeapot, isAPI)
		return
	}

	dupes, err := s.DB.CheckDuplicate(report)
	report.Duplicate = dupes > 0
	if dupes > 0 {
		log.Printf("found %d duplicates of report from: %s", dupes, r.RemoteAddr)
	}

	id, err := s.DB.InsertReport(report)
	if err != nil {
		log.Printf("failed to insert report into database: %v", err)
		s.sendError(w, "", http.StatusInternalServerError, isAPI)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	if err = report.WriteFile(s.Reports, id, name, email); err != nil {
		log.Printf("failed to write file: %d\n", id)
		s.sendError(w, "", http.StatusInternalServerError, isAPI)
		return
	}

	if !report.Duplicate && s.WH != nil {
		go s.WH.Post(name, id, report.Error.Message)
	}

	if isAPI {
		jsonResponse(w, map[string]interface{}{
			"crashId":  id,
			"crashUrl": fmt.Sprintf("https://crash.pmmp.io/view/%d", id),
		})
	} else {
		http.Redirect(w, r, fmt.Sprintf("/view/%d", id), http.StatusMovedPermanently)
	}
}

func (s Submit) sendError(w http.ResponseWriter, message string, status int, isAPI bool) {
	if isAPI {
		w.WriteHeader(status)
		if message == "" {
			message = http.StatusText(status)
		}
		jsonResponse(w, map[string]interface{}{
			"error": message,
		})
	} else {
		s.View.Error(w, message, status)
	}
}

func jsonResponse(w http.ResponseWriter, data map[string]interface{}) {
	json.NewEncoder(w).Encode(data)
}

func ParseMultipartForm(form *multipart.Form) (string, error) {
	var report string
	if reportPaste, ok := form.Value["reportPaste"]; ok && reportPaste[0] != "" {
		report = reportPaste[0]
	} else if reportFile, ok := form.File["reportFile"]; ok && reportFile[0] != nil {
		f, err := reportFile[0].Open()
		if err != nil {
			return "", errors.New("could not open file")
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return "", errors.New("could not read file")
		}
		f.Close()
		report = string(b)
	}

	return report, nil
}
