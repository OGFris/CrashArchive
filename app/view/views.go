package view

import (
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/pmmp/CrashArchive/app/crashreport"
)

const (
	templateExtension = ".html"
	templateLayout    = "layout"
)

type Views map[string]*template.Template

func Load(path string) (Views, error) {
	views := make(Views, 0)
	abs, _ := filepath.Abs(path)

	layoutFiles, err := filepath.Glob(filepath.Join(abs, templateLayout, "*"+templateExtension))
	if err != nil {
		return views, err
	}

	pageFiles, err := filepath.Glob(filepath.Join(abs, "*"+templateExtension))
	if err != nil {
		return views, err
	}

	for _, page := range pageFiles {
		templateFiles := append(layoutFiles, page)
		_, fname := filepath.Split(page)

		name := fname[:len(fname)-len(templateExtension)]
		tmpl, err := template.New(name).Funcs(funcMap).ParseFiles(templateFiles...)
		if err != nil {
			return views, err
		}
		views[name] = tmpl
	}
	return views, nil
}

func (v Views) Execute(w http.ResponseWriter, name string, data interface{}) error {
	if tmpl, ok := v[name]; ok {
		return tmpl.ExecuteTemplate(w, "base.html", data)
	}
	return v.Error(w, "", http.StatusInternalServerError)
}

func (v Views) Error(w http.ResponseWriter, message string, status int) error {
	w.WriteHeader(status)
	if message == "" {
		message = http.StatusText(status)
	}
	return v["error"].ExecuteTemplate(w, "base.html", struct{ Message string }{message})
}

func (v Views) ExecuteList(w http.ResponseWriter, reports []crashreport.Report, url string, id int, start int, total int) {
	cnt := len(reports)

	data := map[string]interface{}{
		"RangeStart": 0,
		"RangeEnd":   start + cnt,
		"ShowCount":  cnt,
		"TotalCount": total,
		"SearchUrl":  url,
		"Data":       reports,
		"PrevPage":   0,
		"NextPage":   0,
	}

	if cnt > 0 {
		data["RangeStart"] = start + 1
	}

	if start > 0 {
		data["PrevPage"] = id - 1
	}

	if start+cnt < total {
		data["NextPage"] = id + 1
	}
	v.Execute(w, "list", data)
}
