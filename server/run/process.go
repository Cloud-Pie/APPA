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
)

// this will be responsible for taking the data in the format
// starting the server
// starting the process
// pushing the file to a storage after the process is done

func GetVMStartScript(s3BucketName string)string{
	var VMStartScript = "#!bin/sh \n"+
		"echo \"setup\"  \n"+
		"apt-get install -y linux-image-extra-$(uname -r) linux-image-extra-virtual  \n"+
		"apt-get update  \n"+
		"apt-get install -y apt-transport-https ca-certificates curl software-properties-common  \n"
// add some more code here
	encodedString:=b64.StdEncoding.EncodeToString([]byte(VMStartScript))

	return encodedString
}

func startTestVM( s3BucketName, testVMType string)  string {

	sessionAWS := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AWSConfig.AwsAccessKeyId, AWSConfig.AwsSecretAccessKey, ""),
		Region:      aws.String(AWSConfig.Region),
	}))

	svc := ec2.New(sessionAWS)
	var allInstancesStarted []Ec2Instances

	input := &ec2.RunInstancesInput{
		BlockDeviceMappings: []*ec2.BlockDeviceMapping{
			{
				DeviceName: aws.String("/dev/sdh"),
				Ebs: &ec2.EbsBlockDevice{
					VolumeSize: aws.Int64(20),
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
						Key:   aws.String("Purpose"),
						Value: aws.String("test"),
					},
				},
			},
		},
		UserData: aws.String(GetVMStartScript(s3BucketName)),
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

func terminateTestVM(instanceId string) {

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




func launchVMandDeploy(s3BucketName , testVMType string ){

	log.Println("Starting a test VM of type ", testVMType, " and running the application")

	startedInstanceId :=startTestVM(s3BucketName, testVMType)
	if( startedInstanceId==""){
		log.Fatal("Cannot start test VM, terminating test start again latter")
		return
	}
	stopChecking := Schedule(func() {
		log.Println("waiting for some time for the VM to start and run app")
		// need to have a mechanism by which I query application and stop checking whether its deployed or not
		getVMPublicIP(startedInstanceId)
	}, 30*time.Second)
	time.Sleep(12 * time.Minute)

	// assuming that it might be finished need to add some check conditions here
	stopChecking <- true
	publicAddress:= getVMPublicIP(startedInstanceId)
	log.Println("Public Ip Address : ",publicAddress )
	log.Println("Starting the App")

	// after some time the VM needs to be stopped after the test is finished
	time.Sleep(15 * time.Minute)


	//savemonitoringDump(publicAddress, backUpDirectoryName)

	log.Println(" Terminating the VM")
	terminateTestVM(startedInstanceId)
}