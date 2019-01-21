package run

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"log"
	b64 "encoding/base64"
	"time"
	"github.com/aws/aws-sdk-go/service/s3"
	"gopkg.in/mgo.v2/bson"
	"strings"
	"fmt"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"strconv"
	"os"
)

// this will be responsible for taking the data in the format
// starting the server
// starting the process
// pushing the file to a storage after the process is done

func createS3Bucket(s3BucketName string) bool{

	sessionAWS := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AWSConfig.AwsAccessKeyId, AWSConfig.AwsSecretAccessKey, ""),
		Region:      aws.String(AWSConfig.Region),
	}))
	// Create S3 service client
	svc := s3.New(sessionAWS)

	input := &s3.CreateBucketInput{
		Bucket: aws.String(s3BucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(AWSConfig.Region),
		},
	}

	result, err := svc.CreateBucket(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				log.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
				return true

			case s3.ErrCodeBucketAlreadyOwnedByYou:
				log.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
				return true
			default:
				log.Println(aerr.Error())
				return false
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
			return false
		}
	}
	log.Println(result)
	return true
}
/*
func getPublicIpTool() string{
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Oops: " + err.Error() + "\n")
		return ""
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Println(ipnet.IP.String() + "\n")
				return ipnet.IP.String()
			}
		}
	}
	return ""
}*/
//function to get the public ip address
func getPublicIpTool() string {

	cmd:="dig +short myip.opendns.com @resolver1.opendns.com"
	wanip:=exe_cmd_output(cmd)
	fmt.Println(wanip)
	wanip = strings.TrimSuffix(wanip, "\n")
	fmt.Println(wanip)
	return string(wanip)
}

func getVMStartScript(gitPath,testName, publicIpTool,test_case, maxTimeSteps,authContents string)string{
	var VMStartScript = `#!bin/sh
apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual 
apt-get update  
apt-get install -y apt-transport-https ca-certificates curl software-properties-common
apt-get --assume-yes install git
apt-get install -y python-pip python-dev build-essential 
apt-get install -y unzip
apt-get -y install awscli
apt-get install -y \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
apt-get update
apt-get install -y docker-ce
curl -XPOST 'http://`+publicIpTool+`:8086/query' --data-urlencode 'q=CREATE DATABASE "`+testName+`"'
pip install awscli --upgrade --user
git clone https://github.com/ansjin/docker-node-monitoring.git
FILE="docker-node-monitoring/local/prometheus/prometheus.yml"
cat <<EOT >> $FILE
remote_write:
  - url: "http://`+publicIpTool+`:8086/api/v1/prom/write?db=`+testName+`&u=root&p=root"
remote_read:
  - url: "http://`+publicIpTool+`:8086/api/v1/prom/read?db=`+testName+`&u=root&p=root"
EOT
cd docker-node-monitoring/local/scripts
sh ./deploy_app.sh
# Define a timestamp function
timestamp() {
  date +"%T"
}
cd /
destdir="/service-account.json"
cat <<EOT >> $destdir
`+authContents+`
EOT
wget https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.tar.gz
tar xfz google-cloud-sdk.tar.gz -C ./
cd google-cloud-sdk 
./install.sh
gcloud auth activate-service-account --key-file=/service-account.json
gsutil cp /service-account.json gs://`+GCEConfig.BucketName+`/
aws configure set aws_access_key_id `+AWSConfig.AwsAccessKeyId+`
aws configure set aws_secret_access_key `+AWSConfig.AwsSecretAccessKey+`
aws configure set default.region `+AWSConfig.Region+`
aws configure set region `+AWSConfig.Region+`
aws s3 cp /service-account.json s3://`+AWSConfig.S3BucketName+`/
git clone `+ gitPath+ `
aws s3 cp s3://boundarydata/Inlet_Data.zip Inlet_Data.zip
unzip Inlet_Data.zip -d Inlet_Data
cp -R Inlet_Data/Inlet_Data/constant/ openfoam/`+ test_case+ `/openfoam_src/code/
cd openfoam/`+ test_case+ `/scripts
sh ./deploy_app.sh
cd /openfoam/`+ test_case+ `/openfoam_src/code/
maxTimeSteps=`+ maxTimeSteps+ `
currentStatus=0
while [ $currentStatus != $maxTimeSteps ]
do
   currentVal=$(ls -td -- */ | head -n 1 | cut -d'/' -f1)
   curl -L "http://`+publicIpTool+`:8080/updateCurrentStatus/`+testName+`/$currentVal"
   currentStatus=$currentVal
   sleep 5m
done
if [ $currentStatus = $maxTimeSteps]
then
	sleep 10m
	new_fileName=/openfoam/`+ test_case+ `/results/`+testName+`.tar.gz
    mv /openfoam/`+ test_case+ `/results/result.tar.gz $new_fileName
	aws s3 cp $new_fileName s3://`+AWSConfig.S3BucketName+`/
else
    echo "some issue with if "
fi
curl -L "http://`+publicIpTool+`:8080/testFinishedTerminateVM/`+testName+`"
`
	encodedString:=b64.StdEncoding.EncodeToString([]byte(VMStartScript))

	return encodedString
}

func startTestVM( gitAppPath, testVMType,testName,test_case,maxTimeSteps string)  string {

	sessionAWS := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AWSConfig.AwsAccessKeyId, AWSConfig.AwsSecretAccessKey, ""),
		Region:      aws.String(AWSConfig.Region),
	}))

	svc := ec2.New(sessionAWS)
	var allInstancesStarted []Ec2Instances
	authContents:=readFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	fmt.Println(authContents)
	input := &ec2.RunInstancesInput{
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sda1"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(40),
				},
			},
		},
		ImageId:      aws.String(GetImageFromRegion(AWSConfig.Region)),
		InstanceType: aws.String(testVMType),
		KeyName:      aws.String(AWSConfig.KeyPairName),
		MaxCount:     aws.Int64(1),
		MinCount:     aws.Int64(1),
		SecurityGroupIds: []*string{
			aws.String(AWSConfig.SecurityGroup),
		},
		SubnetId: aws.String(AWSConfig.SubnetId),
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String("instance"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("Name"),
						Value: aws.String("APPA"),
					},
				},
			},
		},
		UserData: aws.String(getVMStartScript(gitAppPath,testName, AWSConfig.PublicIpServer, test_case,maxTimeSteps,authContents)),
	}

	result, err := svc.RunInstances(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Fatal(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Fatal(err.Error())
		}
		return ""
	}

	for _, instance := range result.Instances {

		oneInstance := Ec2Instances{InstanceId: ValueAssignString(instance.InstanceId, ""),
			ImageId: ValueAssignString(instance.ImageId, ""),
			InstanceType: ValueAssignString(instance.InstanceType, ""),
			LaunchTime: *instance.LaunchTime,
			InstanceState: ValueAssignString(instance.State.Name, ""),
			AvailabilityZone: ValueAssignString(instance.Placement.AvailabilityZone, ""),
			CoreCount: ValueAssignInt64(instance.CpuOptions.CoreCount, 0),
			PublicIpAddress: ValueAssignString(instance.PublicIpAddress, "")}

		allInstancesStarted = append(allInstancesStarted, oneInstance)
	}
	log.Println(allInstancesStarted)
	return allInstancesStarted[0].InstanceId
}

func terminateTestVM(instanceId, csp, zone string) {

	switch csp {

	case "AWS":
		session := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(AWSConfig.AwsAccessKeyId, AWSConfig.AwsSecretAccessKey, ""),
			Region:      aws.String(AWSConfig.Region),
		}))

		svc2 := ec2.New(session)

		var allInstances []string

		input2 := ec2.TerminateInstancesInput{InstanceIds: []*string{
			aws.String(instanceId),
		},}
		result2, er2r := svc2.TerminateInstances(&input2)
		if er2r != nil {
			log.Fatal(er2r)
			return
		}
		for _, instance := range result2.TerminatingInstances {
			allInstances = append(allInstances, ValueAssignString(instance.InstanceId, ""))
		}

		log.Println("Terminate Instances with id: ", allInstances)

	case "GCE":
		deleteAll(instanceId,zone)
	}


}


func getVMPublicIP(startedInstanceId string)  string{
	session := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AWSConfig.AwsAccessKeyId, AWSConfig.AwsSecretAccessKey, ""),
		Region:      aws.String(AWSConfig.Region),
	}))

	svc2 := ec2.New(session)

	var allInstances []Ec2Instances

	input2 := ec2.DescribeInstancesInput{InstanceIds: []*string{
		aws.String(startedInstanceId),
	},}
	result2, er2r := svc2.DescribeInstances(&input2)
	if er2r != nil {
		log.Fatal(er2r)
		return ""
	}
	for _, reservation := range result2.Reservations {
		for _, instance := range reservation.Instances {

			oneInstance := Ec2Instances{InstanceId: ValueAssignString(instance.InstanceId, ""),
				ImageId: ValueAssignString(instance.ImageId, ""),
				InstanceType: ValueAssignString(instance.InstanceType, ""),
				LaunchTime: *instance.LaunchTime,
				InstanceState: ValueAssignString(instance.State.Name, ""),
				AvailabilityZone: ValueAssignString(instance.Placement.AvailabilityZone, ""),
				CoreCount: ValueAssignInt64(instance.CpuOptions.CoreCount, 0),
				PublicIpAddress: ValueAssignString(instance.PublicIpAddress, "")}

			allInstances = append(allInstances, oneInstance)
		}
	}
	log.Println(allInstances[0].PublicIpAddress)
	return allInstances[0].PublicIpAddress
}


func updateTargetFiles(ip, port, typeTarget, fileName string, targetArray []PrometheusTarget ){
	oneTarget:= ip+ ":"+ port
	var targets []string
	targets = append(targets, oneTarget)
	myTarget:= PrometheusTarget{targets,LabelDef{typeTarget, ip} }
	targetArray = append(targetArray, myTarget)
	allTargetsJson, _ := json.Marshal(targetArray)
	err := ioutil.WriteFile(fileName, allTargetsJson, 0666)
	fmt.Printf("%+v", targetArray)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("%+v", targetArray)
	}else {
		fmt.Println("File Written", fileName)
	}
}

func startInstanceAWS( testName, gitAppPath , testVMType,test_case, numCells, numCores,maxTimeSteps string){

	startedInstanceId :=startTestVM(gitAppPath, testVMType, testName,test_case,maxTimeSteps)
	if( startedInstanceId==""){
		log.Fatal("Cannot start test VM, terminating test start again latter")
		return
	}
	mongoSession := GetMongoSession()
	collection := mongoSession.DB(Database).C(Collection_Name)


	AllData := TestInformation{
		TestName			:  	testName,
		BucketName			:  	AWSConfig.S3BucketName,
		InstanceId 			: 	startedInstanceId,
		Region				:  	AWSConfig.Region,
		StartTimestamp		:	time.Now().Unix(),
		NumInstances		:   1,
		InstanceType		:	testVMType,
		GitPath				: 	gitAppPath,
		FileName			: 	testName+".tar.gz",
		NumCells 			: 	numCells,
		NumCores			: 	numCores,
		Phase				:   "Deployment",
		Test_case			:  test_case,
		CSP					: "AWS",
		LastUpdated 		:	time.Now().Unix(),
		CurrentStatus		: "0",
		MaxTimeSteps		: maxTimeSteps,
	}
	if err := collection.Insert(AllData); err != nil {
		log.Fatal("error ", err)
	} else {
		log.Println("#inserted into ", Collection_Name)
	}

	stopChecking := Schedule(func() {
		log.Println("waiting for some time for the VM to start and run app")
		// need to have a mechanism by which I query application and stop checking whether its deployed or not
		getVMPublicIP(startedInstanceId)
	}, 30*time.Second)
	time.Sleep(1 * time.Minute)

	// assuming that it might be finished need to add some check conditions here
	stopChecking <- true
	publicAddress:= getVMPublicIP(startedInstanceId)
	log.Println("Public Ip Address : ",publicAddress )
	log.Println("Starting the App")
	///updateTargetFiles(publicAddress, "9323","docker_remote", "/targets/targets_docker.json", targetsDocker)
	///updateTargetFiles(publicAddress, "9091","pushgateway_remote", "/targets/targets_pushgateway.json", targetsPushGateway)
	//updateTargetFiles(publicAddress, "8080","cadvisor_remote", "/targets/targets_cadvisor.json", targetsCadvisor)
	//updateTargetFiles(publicAddress, "9100","nodeexporter_remote", "/targets/targets_nodeexporter.json", targetsNodeExporter)
	fmt.Println("testname", testName)
	errMongoU := collection.Update(bson.M{"testname": testName}, bson.M{"$set": bson.M{"phase": "Deployed","publicipaddress": publicAddress}})
	if errMongoU != nil {
		log.Fatal("Error : %s", errMongoU)
	}

	defer mongoSession.Close()
}

func startInstanceGCE( testName, gitAppPath , testVMType,test_case, numCells, numCores,zone,maxTimeSteps string){

	startedInstanceId :=createGoogleInstance(gitAppPath, testVMType, testName,test_case,zone,maxTimeSteps)
	if( startedInstanceId==""){
		log.Fatal("Cannot start test VM, terminating test start again latter")
		return
	}
	mongoSession := GetMongoSession()
	collection := mongoSession.DB(Database).C(Collection_Name)


	AllData := TestInformation{
		TestName			:  	testName,
		BucketName			:  	GCEConfig.BucketName,
		InstanceId 			: 	startedInstanceId,
		Region				:  	zone,
		StartTimestamp		:	time.Now().Unix(),
		NumInstances		:   1,
		InstanceType		:	testVMType,
		GitPath				: 	gitAppPath,
		FileName			: 	testName+".tar.gz",
		NumCells 			: 	numCells,
		NumCores			: 	numCores,
		Phase				:   "Deployment",
		Test_case			:  test_case,
		CSP					:  "GCE",
		LastUpdated 		:	time.Now().Unix(),
		CurrentStatus		: "0",
		MaxTimeSteps		: maxTimeSteps,
	}
	if err := collection.Insert(AllData); err != nil {
		log.Fatal("error ", err)
	} else {
		log.Println("#inserted into ", Collection_Name)
	}

	stopChecking := Schedule(func() {
		log.Println("waiting for some time for the VM to start and run app")
		// need to have a mechanism by which I query application and stop checking whether its deployed or not
		getInstanceIp(testName,zone)
	}, 30*time.Second)
	time.Sleep(1 * time.Minute)

	// assuming that it might be finished need to add some check conditions here
	stopChecking <- true
	publicAddress:= getInstanceIp(testName,zone)
	log.Println("Public Ip Address : ",publicAddress )
	log.Println("Starting the App")
	///updateTargetFiles(publicAddress, "9323","docker_remote", "/targets/targets_docker.json", targetsDocker)
	///updateTargetFiles(publicAddress, "9091","pushgateway_remote", "/targets/targets_pushgateway.json", targetsPushGateway)
	//updateTargetFiles(publicAddress, "8080","cadvisor_remote", "/targets/targets_cadvisor.json", targetsCadvisor)
	//updateTargetFiles(publicAddress, "9100","nodeexporter_remote", "/targets/targets_nodeexporter.json", targetsNodeExporter)
	fmt.Println("testname", testName)
	errMongoU := collection.Update(bson.M{"testname": testName}, bson.M{"$set": bson.M{"phase": "Deployed","publicipaddress": publicAddress}})
	if errMongoU != nil {
		log.Fatal("Error : %s", errMongoU)
	}

	defer mongoSession.Close()
}
func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}


func launchVMandDeploy(inputValues InputStruct ){

	log.Println("Starting a test VM of type ", inputValues.InstanceType, " and running the application on cloud ", inputValues.CSP)

	testName:= "appa"+RandStringRunes(10)

	switch inputValues.CSP {

	case "aws":
		startInstanceAWS(testName,inputValues.AppGitPath , inputValues.InstanceType,inputValues.Test_case, inputValues.NumCells, inputValues.NumCores , inputValues.MaxTimeSteps)

	case "gce":
		startInstanceGCE(testName,inputValues.AppGitPath , inputValues.InstanceType,inputValues.Test_case, inputValues.NumCells, inputValues.NumCores, inputValues.Zone,inputValues.MaxTimeSteps )
	default:
		log.Println("Not the correct case")

	}


}

func testFinishedTerminateVM(testName string){
	mongoSession := GetMongoSession()
	collection := mongoSession.DB(Database).C(Collection_Name)
	var testInformation TestInformation
	err :=  collection.Find(bson.M{"testname":testName}).One(&testInformation)
	if err != nil {
		log.Fatal("Db Error : ", err)
		return
	}
	log.Println(" Terminating the VM")
	terminateTestVM(testInformation.InstanceId, testInformation.CSP, testInformation.Region)

	errMonFin := collection.Update(bson.M{"testname": testName}, bson.M{"$set": bson.M{"endtimestamp": time.Now().Unix(),
		"phase": "Completed"}})
	if errMonFin != nil {
		log.Fatal("Error::%s", errMonFin)
	}
	defer mongoSession.Close()
}

func updateTestStatus(testName, currentStatus string){
	mongoSession := GetMongoSession()
	collection := mongoSession.DB(Database).C(Collection_Name)
	var testInformation TestInformation
	err :=  collection.Find(bson.M{"testname":testName}).One(&testInformation)
	if err != nil {
		log.Fatal("Db Error : ", err)
		return
	}
	log.Println(" Updating Test Status")

	if _, err := strconv.Atoi(currentStatus); err == nil {
		errMonFin := collection.Update(bson.M{"testname": testName}, bson.M{"$set": bson.M{"currentstatus":currentStatus,"lastupdated": time.Now().Unix(), }})
		if errMonFin != nil {
			log.Fatal("Error::%s", errMonFin)
		}
	}else{
		errMonFin := collection.Update(bson.M{"testname": testName}, bson.M{"$set": bson.M{"lastupdated": time.Now().Unix(), }})
		if errMonFin != nil {
			log.Fatal("Error::%s", errMonFin)
		}
	}

	defer mongoSession.Close()
}

func listObjectsInBucket() *s3.ListObjectsOutput{

	// TODO: Need to check whether this is required or not or query mongodb for all the test names as done in the next function

	sessionAWS := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AWSConfig.AwsAccessKeyId, AWSConfig.AwsSecretAccessKey, ""),
		Region:      aws.String(AWSConfig.Region),
	}))
	// Create S3 service client
	svc := s3.New(sessionAWS)

	input := &s3.ListObjectsInput{
		Bucket: aws.String(AWSConfig.S3BucketName),
		MaxKeys: aws.Int64(2),
	}

	result, err := svc.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				log.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				log.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			log.Println(err.Error())
		}
		return nil
	}
	// TODO: Need to check this format, currently its directly sent to cli but only names are to be sent to cli
	log.Println(result)
	return result
}

func getAllTestsInformation() []TestInformation{
	var allTestInformation []TestInformation
	mongoSession := GetMongoSession()
	collection := mongoSession.DB(Database).C(Collection_Name)
	err :=  collection.Find(nil).All(&allTestInformation)
	if err != nil {
		log.Fatal("Db Error : ", err)
	}
	return  allTestInformation
}
func getTestInformation(testName string) TestInformation{
	var testInformation TestInformation
	mongoSession := GetMongoSession()
	collection := mongoSession.DB(Database).C(Collection_Name)
	err :=  collection.Find(bson.M{"testname":testName}).One(&testInformation)
	if err != nil {
		log.Fatal("Db Error : ", err)

	}
	return  testInformation
}