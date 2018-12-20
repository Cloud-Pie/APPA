package run

import (
	"os"
	"fmt"
)

func initConfig()  {

	AllInstanceTypes = []string{ "t2.micro", "t2.nano", "t2.small", "t2.medium", "t2.large","t2.xlarge"}

	DefaultRegion = []string{ "us-east-2", "us-east-1", "us-west-1", "eu-central-1", "ap-south-1", "eu-west-1", "ap-southeast-2"}

	DefaultZone = []string{ "us-east-2a","us-east-1a","us-west-1a","eu-central-1a","ap-south-1a","eu-west-1a", "ap-southeast-2a"}

	DefaultAMI = []string{ "ami-5e8bb23b", "ami-759bc50a", "ami-4aa04129", "ami-de8fb135", "ami-188fba77","ami-2a7d75c0", "ami-47c21a25"}


	AWSConfig =  AWSConfigStruct{AwsAccessKeyId: os.Getenv("AWS_KEY"), AwsSecretAccessKey: os.Getenv("AWS_SECRET"),
		Region:  os.Getenv("AWS_DEFAULT_REGION"), KeyPairName:os.Getenv("AWS_KEY_PAIR_NAME"), SubnetId:os.Getenv("AWS_SUBNET_ID"),
		SecurityGroup:os.Getenv("AWS_SECURITY_GROUP"),S3BucketName: os.Getenv("AWS_S3BUCKET_PREFIX")+"appa", PublicIpServer: "SERVER_PUBLIC_IP_ADDRESS"}

		createS3Bucket(AWSConfig.S3BucketName)
	fmt.Println("IP address: ", AWSConfig.PublicIpServer)
}
