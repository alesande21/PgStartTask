package main

import (
	"PgStartTask/server/api"
	"flag"
	"fmt"
	middleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

func main() {
	var port = flag.Int("port", 8080, "Порт для тестироования http сервера")
	flag.Parse()

	swagger, err := api.GetSwagger()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	scriptServer := api.NewScriptServer()

	r := mux.NewRouter()

	r.Use(middleware.OapiRequestValidator(swagger))
	api.HandlerFromMux(scriptServer, r)

	s := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("0.0.0.0:%d", *port),
	}

	log.Fatal(s.ListenAndServe())

}
