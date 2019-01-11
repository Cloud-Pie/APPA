package run

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"time"
	"strconv"
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

func getVMStartUpScript(gitPath,testName, publicIpTool ,test_case string) string {
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
aws configure set aws_access_key_id `+AWSConfig.AwsAccessKeyId+`
aws configure set aws_secret_access_key `+AWSConfig.AwsSecretAccessKey+`
aws configure set default.region `+AWSConfig.Region+`
aws configure set region `+AWSConfig.Region+`
git clone `+ gitPath+ `
aws s3 cp s3://boundarydata/Inlet_Data.zip Inlet_Data.zip
unzip Inlet_Data.zip -d Inlet_Data
cp -R Inlet_Data/Inlet_Data/constant/ openfoam/`+ test_case+ `/openfoam_src/code/
cd openfoam/`+ test_case+ `/scripts
sh ./deploy_app.sh
$file_name = /results/result.tar.gz 
while [ -ne $file_name ]
do
   sleep 5m
done
if [ -e $file_name]
then
	new_fileName=/results/`+testName+`.tar.gz
    mv $file_name $new_fileName
	aws s3 cp $new_fileName s3://`+AWSConfig.S3BucketName+`/
else
    echo "not found"
fi
curl -L "http://`+publicIpTool+`:8080/testFinishedTerminateVM/`+testName+`"
`
	return VMStartScript
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

func createInstance(project,gitAppPath, testVMType,testName,test_case, zone string) string{

	ctx := context.Background()

	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Println(err)
	}else{

		computeService, err := compute.New(c)
		if err != nil {
			log.Println(err)
		}else{

			vmStartscript:=getVMStartUpScript(gitAppPath,testName, AWSConfig.PublicIpServer, test_case)

			rb := &compute.Instance{
				MachineType:"zones/"+zone+"/machineTypes/"+testVMType,
				Name:"appa-server",
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
func createGoogleInstance(gitAppPath, testVMType,testName,test_case,zone string) string {

	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)

	if err != nil {
		log.Println(err)
	}

	// Project ID for this request.
	project := creds.ProjectID
	createNetwork(project)

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

	instanceID := createInstance(project,gitAppPath, testVMType,testName,test_case, zone)

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

				return resp.NetworkInterfaces[0].NetworkIP
			}
		}
	}
	return ""

}