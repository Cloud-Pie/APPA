package run

import (
	"net/http"
	"log"
	"github.com/gorilla/mux"
)

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
// DeployAndRunApplication godoc
// @Summary Start the VM and run the application
// @Description  Start the VM and run the application
// @Tags START_TEST
// @Accept text/html
// @Produce json
// @Param instancetype query string true " host instance type "
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "ok"
// @Failure 404 {string} string "ok"
// @Failure 500 {string} string "ok"
// @Router /deployAndRunApplication/{instancetype}/[get]
// here need to  pass application specifc parameters as well
func DeployAndRunApplication(w http.ResponseWriter, r *http.Request)  {

	log.Println("Entered DeployAndRunApplication function")
	vars := mux.Vars(r)
	log.Println("instancetype=", vars["instancetype"])

	var instanceType = ""
	if(vars["instancetype"]==""){
		instanceType = "t2.large"
	}
	// here need to  pass application specifc parameters as well
	go launchVMandDeploy("s3buckerNameTobeDecided",instanceType)
	w.Write([]byte("The process has started, check the logs for details!!"))
}