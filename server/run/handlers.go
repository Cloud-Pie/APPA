package run

import "net/http"

func Index(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Welcome to APPA Server!!"))
}


func getData(w http.ResponseWriter, r *http.Request) {

	// a call to data handler will be made here
	w.Write([]byte("Data from the cli will be handled here!!"))
}

func getListFiles(w http.ResponseWriter, r *http.Request)  {

	w.Write([]byte("Here the list of all available files will be provoded!!"))
}