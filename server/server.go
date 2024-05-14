package main

import (
	"PgStartTask/db"
	"PgStartTask/server/api"
	"flag"
	"fmt"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strconv"
)

var (
	Ip   = flag.String("ip", api.Localhost, "Set ip address")
	Port = flag.Int("port", api.DefaultPort, "Set instance port")
)

func main() {
	//var port = flag.Int("port", 8080, "Порт для тестироования http сервера")
	flag.Parse()

	swagger, err := api.GetSwagger()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	realDB := db.RealDB{DB: db.ConnectionToDB()}
	defer realDB.DB.Close()

	commandToRun := make(chan api.Command)

	//var app *Application
	scriptServer := api.NewScriptServer(&realDB, commandToRun)

	go api.ControlRunningCommand(scriptServer)

	r := mux.NewRouter()

	r.Use(middleware.OapiRequestValidator(swagger))
	api.HandlerFromMux(scriptServer, r)

	s := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%d", *Ip, *Port),
	}

	log.Printf("Подключнеие установлено -> %s:%s", api.ColorString(api.FgYellow, *Ip), api.ColorString(api.FgYellow, strconv.Itoa(*Port)))

	log.Fatal(s.ListenAndServe())

}
