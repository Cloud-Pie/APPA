package run

import "time"

type Status int
type Task struct {
	ID string
	Status
}

const (
	InQueue Status = iota
	Running
	Completed
	Failed
)

type AWSConfigStruct struct {
	AwsAccessKeyId   		string 		`json:"AwsAccessKeyId"`
	AwsSecretAccessKey   	string 		`json:"AwsSecretAccessKey"`
	Region 					string 		`json:"Region"`
	KeyPairName				string 		`json:"KeyName"`
	SubnetId				string 		`json:"SubnetId"`
	SecurityGroup			string 		`json:"SecurityGroup"`
	S3BucketName			string 		`json:"S3BucketName"`
	PublicIpServer			string 		`json:"PublicIpServer"`
}

type GCEConfigStruct struct {
	BucketName				string 		`json:"BucketName"`
	NetworkName				string 		`json:"NetworkName"`
	PublicIpServer			string 		`json:"PublicIpServer"`
}

type Ec2Instances struct {
	InstanceId       		string 		`json:"InstanceId"`
	InstanceState     		string 		`json:"InstanceState"`
	AvailabilityZone      	string 		`json:"AvailabilityZone"`
	PublicIpAddress      	string 		`json:"PublicIpAddress"`
	InstanceType 			string 		`json:"InstanceType"`
	ImageId 				string 		`json:"ImageId"`
	CoreCount				int64 		`json:"CoreCount"`
	LaunchTime				time.Time 	`json:"LaunchTime"`
}
type InputStruct struct {
	AppGitPath 				string 		`json:"AppGitPath"`
	InstanceType 			string		`json:"InstanceType"`
	Test_case 				string		`json:"Test_case"`
	NumCores 				string		`json:"NumCores"`
	NumCells 				string		`json:"NumCells"`
	CSP  				    string		`json:"CSP"`
	Zone  				    string		`json:"Zone"`

}
type LabelDef struct {
	Type   			string 	 `json:"type"`
	Ip   			string   `json:"ip"`
}
type PrometheusTarget struct {
	Targets  []string `json:"targets"`
	Labels   LabelDef `json:"labels"`
}

var targetsDocker 						[]PrometheusTarget
var targetsCadvisor 					[]PrometheusTarget
var targetsNodeExporter					[]PrometheusTarget
var targetsPushGateway					[]PrometheusTarget
var AWSConfig 							AWSConfigStruct
var GCEConfig 							GCEConfigStruct
var AllInstanceTypes 					[]string
var DefaultRegion 						[]string
var DefaultZone							[]string
var DefaultAMI 							[]string