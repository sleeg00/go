package main

import (
	"log"

	"github.com/sleeg00/LetsGo/prolog/internal/server"
)

func main() {

	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}
