package main // import "github.com/anyaddres/supermann/modules"

import (
	"log"
	"net/http"
	"runtime"

	"github.com/anyaddres/supermann/api"
)

const (
	// Port ...
	Port = ":8080"
)

func main() {
	log.Println("Starting logins identification server...")
	log.Printf("Go version %s", runtime.Version())
	mux := http.NewServeMux()

	appServer := api.NewServer()
	appServer.DbPing()
	defer appServer.ServerCleanup()
	mux.Handle("/api/identifylogins/", appServer)
	err := http.ListenAndServe(Port, mux)
	if err != nil {
		log.Fatal(err)
	}
}
