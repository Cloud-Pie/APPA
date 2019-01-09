package run

import (
	"net/http"
	"text/template"
	"log"
	"github.com/gorilla/mux"
	"encoding/json"
)

func Index(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.New("").ParseFiles("templates/index.html", "templates/base.html")
	// check your err
	if(err!=nil){
		log.Println("err")
	}else{
		err = tmpl.ExecuteTemplate(w, "base", "")
	}
}
func LogsPage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").ParseFiles("templates/logs.html", "templates/base.html")
	// check your err
	if(err!=nil){
		log.Println("err")
	}else{
		err = tmpl.ExecuteTemplate(w, "base", "")
	}
}
func AllTestsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("").ParseFiles("templates/all_tests.html", "templates/base.html")
	// check your err
	if(err!=nil){
		log.Println("err")
	}else{
		err = tmpl.ExecuteTemplate(w, "base", "")
	}
}

func ConductTestAWSHandler(w http.ResponseWriter, r *http.Request) {


	tmpl, err := template.New("").ParseFiles("templates/conduct_test_aws.html", "templates/base.html")
	// check your err
	if(err!=nil){
		log.Println("err")
	}else{
		err = tmpl.ExecuteTemplate(w, "base", "")
	}


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
// @Param body body run.InputStruct true "..."
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "ok"
// @Failure 404 {string} string "ok"
// @Failure 500 {string} string "ok"
// @Router /deployAndRunApplication/[POST]
func DeployAndRunApplication(w http.ResponseWriter, r *http.Request)  {

	log.Println("Entered DeployAndRunApplication function")
	decoder := json.NewDecoder(r.Body)
	var inputValues InputStruct
	err := decoder.Decode(&inputValues)
	if err != nil {
		log.Fatal(err)
		w.Write([]byte("The process has ended due to an error, check the logs for details!!"))
	}else{
		//TODO: DO error checks whether the fields are empty or nor
		//TODO: Need to add monitoring agent for VM, containers to collect and store all data for later analysis
		go launchVMandDeploy(inputValues.AppGitPath,inputValues.InstanceType, inputValues.Test_case, inputValues.NumCells, inputValues.NumCores)
		w.Write([]byte("The process has started, check the logs for details!!"))
	}
}

// TestFinishedTerminateVM godoc
// @Summary Terminate the VM when the test is finished
// @Description  Terminate the VM when the test is finished
// @Tags COMPLETED_TEST
// @Accept text/html
// @Produce json
// @Param testname query string true " testname "
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "ok"
// @Failure 404 {string} string "ok"
// @Failure 500 {string} string "ok"
// @Router /testFinishedTerminateVM/{testname}/[get]
func TestFinishedTerminateVM(w http.ResponseWriter, r *http.Request)  {

	vars := mux.Vars(r)
	log.Println("testname=", vars["testname"])

	go testFinishedTerminateVM(vars["testname"])

	w.Write([]byte("VM will be terminated"))
}