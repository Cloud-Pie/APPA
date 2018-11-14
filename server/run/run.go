package run

import (
	"log"
	"net/http"
)

func RunServer()  {
	initConfig()
	router := NewRouter()
	log.Println(http.ListenAndServe(":8080", router))
	log.Println("Server started on port 8080")
}
