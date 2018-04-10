package router

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"github.com/pmmp/CrashArchive/app/handler"
)

func New(c *handler.Common) *chi.Mux {
	r := chi.NewRouter()

	staticDirs := []string{"/css", "/js", "/fonts"}
	workDir, _ := os.Getwd()
	for _, v := range staticDirs {
		dir := filepath.Join(workDir, "static", v[1:])
		FileServer(r, v, http.Dir(dir))
	}

	r.Route("/", func(r chi.Router) {
		r.Use(RealIP)
		r.Use(middleware.Logger)

		r.NotFound(func(w http.ResponseWriter, r *http.Request) {
			c.View.Error(w, "", http.StatusNotFound)
		})

		r.Get("/", handler.Home{c}.Get)
		r.Get("/list", handler.List{c}.Get)
		r.Get("/view/{reportID}", handler.View{c}.Get)
		r.Get("/download/{reportID}", handler.Download{c}.Get)

		r.Mount("/search", handler.Search{c}.Routes())
		r.Mount("/submit", handler.Submit{c}.Routes())
	})
	return r
}

func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, ":*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
