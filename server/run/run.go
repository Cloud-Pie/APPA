package run

import (
	"log"
	"net/http"
)

func runServer()  {
	router := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", router))
	log.Fatal("Server started on port 8080")
}
