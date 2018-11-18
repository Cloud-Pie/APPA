package run

import (
	"net/http"
	"log"
	"github.com/gorilla/mux"
	"encoding/json"
)

func Index(w http.ResponseWriter, r *http.Request) {

	w.Write([]byte("Welcome to APPA Server!!"))
}


func DownloadData(w http.ResponseWriter, r *http.Request) {

	// TODO: GIve the download link to download the file based upon the name
	// a call to data handler will be made here
	w.Write([]byte("Data from the cli will be handled here!!"))
}

func ListAllStoredFiles(w http.ResponseWriter, r *http.Request)  {

	//Here the list of all available files will be provided!!
	res:=listObjectsInBucket()

	b, err := json.Marshal(res)
	if err != nil {
		log.Println("error:", err)
	}
	w.Write([]byte(b))
}
// GetAllTestsInformation godoc
// @Summary Get all the tests information
// @Description  Get all the tests information
// @Tags COMPLETED_TEST
// @Accept text/html
// @Produce json
// @Success 200 {array} run.TestInformation ""
// @Failure 400 {string} string "ok"
// @Failure 404 {string} string "ok"
// @Failure 500 {string} string "ok"
// @Router /getAllTestsInformation/[get]
func GetAllTestsInformation(w http.ResponseWriter, r *http.Request)  {

	res:=getAllTestsInformation()

	b, err := json.Marshal(res)
	if err != nil {
		log.Println("error:", err)
	}
	w.Write([]byte(b))
}
// GetTestInformation godoc
// @Summary Get a test information
// @Description  Get a test information
// @Tags COMPLETED_TEST
// @Accept text/html
// @Produce json
// @Param testname query string true " testname "
// @Success 200 {object} run.TestInformation ""
// @Failure 400 {string} string "ok"
// @Failure 404 {string} string "ok"
// @Failure 500 {string} string "ok"
// @Router /getTestInformation/{testname}/[get]
func GetTestInformation(w http.ResponseWriter, r *http.Request)  {

	vars := mux.Vars(r)
	log.Println("testname=", vars["testname"])

	res:=getTestInformation(vars["testname"])

	b, err := json.Marshal(res)
	if err != nil {
		log.Println("error:", err)
	}
	w.Write([]byte(b))
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

	// TODO: Convert this into post query with git path and instance type as params

	log.Println("Entered DeployAndRunApplication function")
	vars := mux.Vars(r)
	log.Println("instancetype=", vars["instancetype"])

	var instanceType = ""
	if(vars["instancetype"]==""){
		instanceType = "t2.large"
	}
	// here need to  pass application specifc parameters as well
	go launchVMandDeploy("git path to be provided",instanceType)
	w.Write([]byte("The process has started, check the logs for details!!"))
}