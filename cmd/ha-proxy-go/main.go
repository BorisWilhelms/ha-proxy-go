package main

import (
	"BorisWilhelms/ha-proxy-go/internal/server"
	"BorisWilhelms/ha-proxy-go/pkg/ha"
	"errors"
	"log"
	"os"
	"text/template"

	"github.com/spf13/viper"
)

var (
	homeassistant ha.HomeAssistant
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

	homeassistant = ha.HomeAssistant{
		BaseUrl:     viper.GetString("BASE_URL"),
		AccessToken: viper.GetString("ACCESS_TOKEN"),
	}

	server := server.Server{
		Homeassistant: homeassistant,
		Automations:   viper.GetStringSlice("automations"),
		Templates:     template.Must(template.ParseGlob("web/template/*")),
	}

	addr := viper.GetString("LISTEN")
	log.Println("Listening on:", addr)
	server.Listen(addr)
}
