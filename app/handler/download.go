package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/pmmp/CrashArchive/app/crashreport"
	"github.com/pmmp/CrashArchive/app/template"
)

func DownloadGet(w http.ResponseWriter, r *http.Request) {
	reportID, err := strconv.Atoi(chi.URLParam(r, "reportID"))
	if err != nil {
		template.ErrorTemplate(w, "Please specify a report", http.StatusNotFound)
		return
	}

	zlibBytes, err := crashreport.ReadRawFile(int64(reportID))
	if err != nil {
		template.ErrorTemplate(w, "Report not found", http.StatusNotFound)
		return
	}

	reportBytes := crashreport.WriteZlibDataToCrashLog(zlibBytes)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.log", reportID))
	w.Header().Set("Content-Length", strconv.Itoa(len(reportBytes)))
	w.Write([]byte(reportBytes))
}
