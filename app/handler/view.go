package handler

import (
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi"

	"github.com/pmmp/CrashArchive/app/crashreport"
)

type View struct{ *Common }

func (v View) Get(w http.ResponseWriter, r *http.Request) {
	reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
	if err != nil {
		v.View.Error(w, "Please specify a report", http.StatusNotFound)
		return
	}
	report, jsonData, err := crashreport.ReadFile(v.Reports, int64(reportID))
	if err != nil {
		v.View.Error(w, "Report not found", http.StatusNotFound)
		return
	}

	data := make(map[string]interface{})
	data["Report"] = report
	data["Name"] = clean(jsonData["name"].(string))
	data["PocketMineVersion"] = report.Version.Get(true)
	data["AttachedIssue"] = "None"
	data["ReportID"] = reportID

	v.View.Execute(w, "view", data)
}

var cleanRE = regexp.MustCompile(`[^A-Za-z0-9_\-\.\,\;\:/\#\(\)\\ ]`)

func clean(v string) string {
	return cleanRE.ReplaceAllString(v, "")
}
