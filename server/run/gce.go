package run

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
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

func createGoogleInstance() {
	ctx := context.Background()

	creds, err := google.FindDefaultCredentials(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}


	c, err := google.DefaultClient(ctx, compute.CloudPlatformScope)
	if err != nil {
		log.Fatal(err)
	}

	computeService, err := compute.New(c)
	if err != nil {
		log.Fatal(err)
	}


	// Project ID for this request.
	project := creds.ProjectID // TODO: Update placeholder value.

	// The name of the zone for this request.
	zone := "us-central1-a" // TODO: Update placeholder value.

	var vmNetworkInterfaces []*compute.NetworkInterface
	vmNetworkInterfaces  = append(vmNetworkInterfaces,  &compute.NetworkInterface{Network:"global/networks/default"})

	var vmDisks []*compute.AttachedDisk
	vmDisks  = append(vmDisks,  &compute.AttachedDisk{InitializeParams: &compute.AttachedDiskInitializeParams{Description:"instance disk for appa server",DiskSizeGb:50, SourceImage:"family/ubuntu-1804-lts"}})

	rb := &compute.Instance{
		MachineType:"zones/us-central1-a/machineTypes/n1-standard-1",
		Name:"appa-server",
		NetworkInterfaces:vmNetworkInterfaces,
		Disks:vmDisks,
	// TODO: Add desired fields of the request body.
	}

	resp, err := computeService.Instances.Insert(project, zone, rb).Context(ctx).Do()
	if err != nil {
		log.Fatal(err)
	}

	// TODO: Change code below to process the `resp` object:
	fmt.Printf("%#v\n", resp)
}