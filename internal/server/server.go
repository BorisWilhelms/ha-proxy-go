package server

import (
	"BorisWilhelms/ha-proxy-go/pkg/ha"
	"io/fs"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	Homeassistant ha.HomeAssistant
	Automations   []string
	Templates     *template.Template
	Static        fs.FS
}

func (server Server) Listen(addr string) {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Route("/{automation}", func(r chi.Router) {
		r.Use(server.automationCtx)
		r.Get("/", server.getAutomation)
		r.Post("/", server.postAutomation)
	})

	fs := http.FileServer(http.FS(server.Static))
	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	http.ListenAndServe(addr, router)
}
