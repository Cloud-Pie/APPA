package main

// TODO: implement /runTask
//TODO: implement /download
//TODO: implement /list

import (
	"encoding/json"
	"net/http"

	"github.com/Cloud-Pie/APPA/server/run"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func runTask(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func download(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func listTasks(w http.ResponseWriter, r *http.Request) {
	tasks := []run.Task{{
		1, run.Running,
	}, {2, run.Failed}, {3, run.Completed}, {4, run.InQueue}, {5, run.Running}, {6, run.Completed}}
	js, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}
func main() {
	http.HandleFunc("/runTask", runTask)

	http.HandleFunc("/list", listTasks)

	http.HandleFunc("/download", download)

	http.HandleFunc("/ping", ping)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
