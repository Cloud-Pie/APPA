package main

// TODO: implement /runTask
//TODO: implement /download
//TODO: implement /list
import (
	"encoding/json"
	"net/http"
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
	tasks := []Task{{
		1, Running,
	}, {2, Failed}, {3, Completed}, {4, InQueue}, {5, Running}, {6, Completed}}
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
