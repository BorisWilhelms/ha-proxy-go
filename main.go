package main

import (
	homeassistant "BorisWilhelms/ha-proxy-go/ha"
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/viper"
)

var (
	ha homeassistant.HomeAssistant
)

func main() {
	viper.SetEnvPrefix("HA_PROXY")
	viper.SetDefault("LISTEN", ":3000")
	viper.SetConfigType("toml")
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()

	if _, err := os.Stat(viper.GetString("ACCESS_TOKEN_FILE")); errors.Is(err, os.ErrNotExist) {
		log.Fatalln("ACCESS_TOKEN_FILE does not exists:", viper.GetString("ACCESS_TOKEN_FILE"))
	}

	data, err := os.ReadFile(viper.GetString("ACCESS_TOKEN_FILE"))
	if err != nil {
		panic(err)
	}
	viper.Set("ACCESS_TOKEN", string(data))

	fs := http.FileServer(http.Dir("wwwroot"))

	ha = homeassistant.HomeAssistant{
		BaseUrl:     viper.GetString("BASE_URL"),
		AccessToken: viper.GetString("ACCESS_TOKEN")}

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Route("/{automation}", func(r chi.Router) {
		r.Use(automationCtx)
		r.Get("/", getAutomation)
		r.Post("/", postAutomation)
	})

	router.Handle("/static/*", http.StripPrefix("/static/", fs))

	addr := viper.GetString("LISTEN")
	log.Println("Listening on:", addr)
	http.ListenAndServe(addr, router)
}

func automationCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		automation := chi.URLParam(r, "automation")
		automations := viper.GetStringSlice("AUTOMATIONS")

		if !contains(automations, automation) {
			log.Println("Automation not found in allowed autiomations:", automation)
			http.Error(w, "automation not found", http.StatusNotFound)
			return
		}
		e := ha.GetState(automation)
		ctx := context.WithValue(r.Context(), "automation", e)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getAutomation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	automation, ok := ctx.Value("automation").(homeassistant.Entity)
	if !ok {
		http.Error(w, "automation not found", http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	model := indexModel{Name: automation.FriendlyName()}
	tmpl.Execute(w, model)
}

func postAutomation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	automation, ok := ctx.Value("automation").(homeassistant.Entity)
	if !ok {
		http.Error(w, "automation not found", http.StatusNotFound)
		return
	}

	ha.CallService("automation", "trigger", automation.Entity_id)
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	model := indexModel{Name: automation.FriendlyName(), Run: true}
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

type indexModel struct {
	Name string
	Run  bool
}
