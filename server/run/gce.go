package run

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"time"
	"strconv"
	"io/ioutil"
	"os"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// BEFORE RUNNING:
// ---------------
// 1. If not already done, enable the Compute Engine API
//    and check the quota for your project at
//    https://console.developers.google.com/apis/api/compute
// 2. This sample uses Application Default Credentials for authentication.
//    If not already done, install the gcloud CLI from
//    https://cloud.google.com/sdk/ and run
//    `gcloud beta auth application-default login`.
//    For more information, see
//    https://developers.google.com/identity/protocols/application-default-credentials
// 3. Install and update the Go dependencies by running `go get -u` in the
//    project directory.

func getVMStartUpScript(gitPath,testName, publicIpTool ,test_case, maxTimeSteps,authContents string) string {
	var VMStartScript = `#!bin/bash
apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual 
apt-get update  
apt-get install -y apt-transport-https ca-certificates curl software-properties-common
apt-get --assume-yes install git
apt-get install -y python-pip python-dev build-essential 
apt-get install -y unzip
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
destdir="/service-account.json"
cat <<EOT >> $destdir
`+authContents+`
EOT
cd /
wget https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.tar.gz
tar xfz google-cloud-sdk.tar.gz -C ./
cd google-cloud-sdk 
./install.sh
source /google-cloud-sdk/completion.bash.inc
source /google-cloud-sdk/path.bash.inc
gcloud auth activate-service-account --key-file=/service-account.json
cd /
git clone `+ gitPath+ `
sudo rm -f /etc/boto.cfg
export BOTO_CONFIG=/dev/null
gsutil cp gs://boundarydata/Inlet_Data.zip Inlet_Data.zip
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
	export BOTO_CONFIG=/dev/null
	gsutil cp $new_fileName gs://`+GCEConfig.BucketName+`/
	curl -L "http://`+publicIpTool+`:8080/testFinishedTerminateVM/`+testName+`"
else
    echo "some issue with if "
fi
`
	return VMStartScript
}

func readFile(filePath string) string{

	dat, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println(err)
	}
	return string(dat)
}

func createNetwork(project string ){
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Println(err)
	}else{
		computeService, err := compute.New(c)
		if err != nil {
			log.Println(err)
		}else{

			resp, err := computeService.Networks.Get(project, GCEConfig.NetworkName).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			}

			// TODO: Change code below to process the `resp` object:
			fmt.Printf("%#v\n", resp)

			if(resp!=nil){

				// already exists
			}else{

				rbNetwork := &compute.Network{
					RoutingConfig: &compute.NetworkRoutingConfig{RoutingMode:"GLOBAL"},
					Name:GCEConfig.NetworkName,
					Description:"network for appa",
					AutoCreateSubnetworks:true,
					// TODO: Add desired fields of the request body.
				}

				respNetwork, err := computeService.Networks.Insert(project, rbNetwork).Context(ctx).Do()
				if err != nil {
					log.Println(err)
				}else{
					// TODO: Change code below to process the `resp` object:
					fmt.Printf("%#v\n", respNetwork)
				}

			}

		}
	}
}

func createBucket(project string){
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}else {

		// Sets the name for the new bucket.
		bucketName := GCEConfig.BucketName

		// Creates a Bucket instance.
		bucket := client.Bucket(bucketName)

		// Creates the new bucket.
		if err := bucket.Create(ctx, project, nil); err != nil {
			log.Println("Failed to create bucket: %v", err)
		}

		fmt.Printf("Bucket %v created.\n", bucketName)
	}

}

func listObjects() []string{

	var objectNames []string
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Println("Failed to create client: %v", err)
	}else {
		// Sets the name for the new bucket.
		bucketName := GCEConfig.BucketName
		it := client.Bucket(bucketName).Objects(ctx, nil)
		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				log.Println(err)
			}else{
				//fmt.Fprintln(w, attrs.Name)
				objectNames = append(objectNames, attrs.Name)
			}
		}
	}
	return objectNames
}

func downloadObject(object string){

	fileName := "/app/assets/"+object
	file, err := os.Create(fileName)
	if err != nil {
		log.Println("Unable to open file %q, %v", err)
	}else {
		defer file.Close()
		ctx := context.Background()
		client, err := storage.NewClient(ctx)
		if err != nil {
			log.Println("Failed to create client: %v", err)
		}else {
			// Sets the name for the new bucket.
			bucketName := GCEConfig.BucketName
			rc, err := client.Bucket(bucketName).Object(object).NewReader(ctx)
			if err != nil {
				log.Println(err)
			}else{
				defer rc.Close()
				data, err := ioutil.ReadAll(rc)
				if err != nil {
					log.Println(err)
				}else {
					n2, err := file.Write(data)
					if err != nil {
						log.Println(err)
					}
					fmt.Printf("wrote %d bytes\n", n2)
				}
			}
		}
	}
}


func addFirewallConfig(project string){
	ctx := context.Background()
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Println(err)
	}else{
		computeService, err := compute.New(c)
		if err != nil {
			log.Println(err)
		}else{

			rbFirewall := &compute.Firewall{
				Allowed: []*compute.FirewallAllowed{&compute.FirewallAllowed{IPProtocol:"all"}, {IPProtocol:"tcp", Ports:[]string{"80","8080"}}},
				Description: "Allowed all traffic",
				Direction: "INGRESS",
				Name:"allow-all",
				Network:"projects/"+project +"/global/networks/"+GCEConfig.NetworkName,
				// TODO: Add desired fields of the request body.
			}

			respFirewall, err := computeService.Firewalls.Insert(project, rbFirewall).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			}else{
				// TODO: Change code below to process the `resp` object:
				fmt.Printf("%#v\n", respFirewall)
			}

		}
	}

}

func createInstance(project,gitAppPath, testVMType,testName,test_case, zone,maxTimeSteps string) string{

	ctx := context.Background()

	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Println(err)
	}else{

		computeService, err := compute.New(c)
		if err != nil {
			log.Println(err)
		}else{

			authContents:=readFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
			//fmt.Println(authContents)
			vmStartscript:=getVMStartUpScript(gitAppPath,testName, AWSConfig.PublicIpServer, test_case,maxTimeSteps,authContents )

			rb := &compute.Instance{
				MachineType:"zones/"+zone+"/machineTypes/"+testVMType,
				Name:testName,
				NetworkInterfaces:[]*compute.NetworkInterface{&compute.NetworkInterface{AccessConfigs: []*compute.AccessConfig{&compute.AccessConfig{Name:"External NAT", NetworkTier:"STANDARD"}},
					Network:"projects/"+project +"/global/networks/"+GCEConfig.NetworkName}},
				Disks:[]*compute.AttachedDisk{&compute.AttachedDisk{ AutoDelete:true, Boot: true, InitializeParams: &compute.AttachedDiskInitializeParams{Description:"instance disk for appa server",
					DiskSizeGb:50, SourceImage:"projects/ubuntu-os-cloud/global/images/family/ubuntu-1804-lts"}}},
				Metadata:&compute.Metadata{Items:[]*compute.MetadataItems{&compute.MetadataItems{Key:"startup-script",Value: &vmStartscript}}},
				// TODO: Add desired fields of the request body.
			}

			resp, err := computeService.Instances.Insert(project, zone, rb).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			}else{
				// TODO: Change code below to process the `resp` object:
				fmt.Printf("%#v\n", resp)
				return strconv.FormatUint(resp.Id, 10)
			}
		}
	}
	return ""
}

func deleteNetwork(){

	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)

	if err != nil {
		log.Println(err)
	}

	// Project ID for this request.
	project := creds.ProjectID

	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Println(err)
	}else{

		computeService, err := compute.New(c)
		if err != nil {
			log.Println(err)
		}else{

			resp, err := computeService.Networks.Get(project, GCEConfig.NetworkName).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			}else{

				// TODO: Change code below to process the `resp` object:
				fmt.Printf("%#v\n", resp)

				if(resp!=nil){

					// already exists
					respNetwork, err := computeService.Networks.Delete(project, GCEConfig.NetworkName).Context(ctx).Do()
					if err != nil {
						log.Println(err)
					}

					// TODO: Change code below to process the `resp` object:
					fmt.Printf("%#v\n", respNetwork)
				}else{

					// no need to do anything
				}
			}
		}
	}
}
func deleteFirewall(){
	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)

	if err != nil {
		log.Println(err)
	}

	// Project ID for this request.
	project := creds.ProjectID
	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Println(err)
	}else{

		computeService, err := compute.New(c)
		if err != nil {
			log.Println(err)
		}else{

			firewall_name:="allow-all"

			respFirewall, err := computeService.Firewalls.Delete(project, firewall_name).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			}else{

				// TODO: Change code below to process the `resp` object:
				fmt.Printf("%#v\n", respFirewall)
			}
		}
	}
}

func deleteInstance(instanceId,zone string){

	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)

	if err != nil {
		log.Println(err)
	}

	// Project ID for this request.
	project := creds.ProjectID

	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Println(err)
	}else{

		computeService, err := compute.New(c)
		if err != nil {
			log.Println(err)
		}else{
			resp, err := computeService.Instances.Delete(project,zone, instanceId).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			}else{
				// TODO: Change code below to process the `resp` object:
				fmt.Printf("%#v\n", resp)
			}
		}
	}
}

func deleteAll(instanceId,zone string)  {
	deleteNetwork()
	deleteFirewall()
	deleteInstance(instanceId,zone)
}
func createGoogleInstance(gitAppPath, testVMType,testName,test_case,zone,maxTimeSteps string) string {

	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)

	if err != nil {
		log.Println(err)
	}

	// Project ID for this request.
	project := creds.ProjectID
	createNetwork(project)
	createBucket(project)

	stopChecking := Schedule(func() {
		log.Println("waiting for some time for the network to become ready")
		// need to have a mechanism by which I query application and stop checking whether its deployed or not
	}, 30*time.Second)
	time.Sleep(2 * time.Minute)

	// assuming that it might be finished need to add some check conditions here
	stopChecking <- true


	addFirewallConfig(project)

	stopChecking2 := Schedule(func() {
		log.Println("waiting for some time for the network to become ready")
		// need to have a mechanism by which I query application and stop checking whether its deployed or not
	}, 30*time.Second)
	time.Sleep(2 * time.Minute)

	// assuming that it might be finished need to add some check conditions here
	stopChecking2 <- true

	instanceID := createInstance(project,gitAppPath, testVMType,testName,test_case, zone,maxTimeSteps)

	return instanceID
	//time.Sleep(3 * time.Minute)

	//deleteAll(project)
}

func getInstanceIp(instanceId,zone string) string{
	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)

	if err != nil {
		log.Println(err)
	}

	// Project ID for this request.
	project := creds.ProjectID

	// The name of the zone for this request.
	//zone := "us-central1-a"

	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Println(err)
	}else{

		computeService, err := compute.New(c)
		if err != nil {
			log.Println(err)
		}else{
			resp, err := computeService.Instances.Get(project, zone, instanceId).Context(ctx).Do()
			if err != nil {
				log.Println(err)
			}else{
				// TODO: Change code below to process the `resp` object:
				fmt.Printf("%#v\n", resp)

				return resp.NetworkInterfaces[0].AccessConfigs[0].NatIP
			}
		}
	}
	return ""

}