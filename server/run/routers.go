package run
import (
	"net/http"
	"github.com/gorilla/mux"
	"strings"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}
	// Declare the static file directory and point it to the directory we just made
	staticFileDirectory := http.Dir("./assets/")
	// Declare the handler, that routes requests to their respective filename.
	// The fileserver is wrapped in the `stripPrefix` method, because we want to
	// remove the "/assets/" prefix when looking for files.
	// For example, if we type "/assets/index.html" in our browser, the file server
	// will look for only "index.html" inside the directory declared above.
	// If we did not strip the prefix, the file server would look for "./assets/assets/index.html", and yield an error
	staticFileHandler := http.StripPrefix("/assets/", http.FileServer(staticFileDirectory))
	// The "PathPrefix" method acts as a matcher, and matches all routes starting
	// with "/assets/", instead of the absolute route itself
	router.PathPrefix("/assets/").Handler(staticFileHandler).Methods("GET")

	return router
}

var routes = Routes{
	Route{
		"Index",
		strings.ToUpper("Get"),
		"/",
		Index,
	},
	Route{
		"DownloadData",
		strings.ToUpper("Get"),
		"/downloadData",
		DownloadData,
	},
	Route{
		"ListAllStoredFiles",
		strings.ToUpper("Get"),
		"/listAllStoredFiles",
		ListAllStoredFiles,
	},
	Route{
		"GetAllTestsInformation",
		strings.ToUpper("Get"),
		"/getAllTestsInformation",
		GetAllTestsInformation,
	},
	Route{
		"GetTestInformation",
		strings.ToUpper("Get"),
		"/getTestInformation/testname",
		GetTestInformation,
	},
	Route{
		"DeployAndRunApplication",
		strings.ToUpper("Post"),
		"/deployAndRunApplication/",
		DeployAndRunApplication,
	},

	Route{
		"TestFinishedTerminateVM",
		strings.ToUpper("Get"),
		"/testFinishedTerminateVM/testname/",
		TestFinishedTerminateVM,
	},

}
