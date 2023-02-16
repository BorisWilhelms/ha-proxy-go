package server

import (
	"BorisWilhelms/ha-proxy-go/pkg/ha"
	"context"
	"io"
	"log"
	"net/http"
	"text/template"

	"github.com/go-chi/chi/v5"
)

func (server Server) automationCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		automation := chi.URLParam(r, "automation")

		if !contains(server.Automations, automation) {
			log.Println("Automation not found in allowed autiomations:", automation)
			http.Error(w, "automation not found", http.StatusNotFound)
			return
		}
		e := server.Homeassistant.GetState(automation)
		ctx := context.WithValue(r.Context(), "automation", e)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (server Server) getAutomation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	automation, ok := ctx.Value("automation").(ha.Entity)
	if !ok {
		http.Error(w, "automation not found", http.StatusNotFound)
		return
	}

	renderIndex(w, indexModel{Name: automation.FriendlyName()})
}

func (server Server) postAutomation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	automation, ok := ctx.Value("automation").(ha.Entity)
	if !ok {
		http.Error(w, "automation not found", http.StatusNotFound)
		return
	}

	server.Homeassistant.CallService("automation", "trigger", automation.Entity_id)
	renderIndex(w, indexModel{Name: automation.FriendlyName(), Run: true})
}

func renderIndex(w io.Writer, model any) {
	tmpl := template.Must(template.ParseFiles("web/template/index.html"))
	tmpl.Execute(w, model)
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
