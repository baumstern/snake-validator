package main

import (
	"log"
	"net/http"
	"validator/server"
)

func main() {
	http.Handle("/new", server.AppHandler(server.NewGame))
	http.Handle("/validate", server.AppHandler(server.Validate))
	
	log.Fatal(http.ListenAndServe(":8080", nil))
}
