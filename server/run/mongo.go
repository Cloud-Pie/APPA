package run

import (
	"gopkg.in/mgo.v2"
	"log"
	"os"
)

var Host = []string{
"mongodb",
}

const (
	Database   			= "APPA"
	Collection_Name		= "ALL_TESTS_CONDUCTED_INFO"
)
var mgoSession   *mgo.Session

type TestInformation struct {
	TestName 				string `json:"TestName"`
	S3BucketName			string `json:"S3BucketName"`
	InstanceId 				string `json:"InstanceId"`
	AWSRegion				string `json:"AWSRegion"`
	StartTimestamp     		int64 `json:"StartTimestamp"`
	NumInstances          	int64 `json:"NumInstances"`
	InstanceType   			string `json:"InstanceType"`
	GitPath            		string `json:"GitPath"`
	Phase              		string `json:"Phase"`
	S3FileName				string `json:"S3FileName"`
	EndTimestamp     		int64 `json:"EndTimestamp"`
}


// Creates a new session if mgoSession is nil i.e there is no active mongo session.
//If there is an active mongo session it will return a Clone
func GetMongoSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.DialWithInfo(&mgo.DialInfo{
			Addrs: Host,
			 Username: os.Getenv("MONGODB_USER"),
			 Password: os.Getenv("MONGODB_PASS"),
			 //Database: os.Getenv("MONGO_INITDB_DATABASE"),
			// DialServer: func(addr *mgo.ServerAddr) (net.Conn, error) {
			// 	return tls.Dial("tcp", addr.String(), &tls.Config{})
			// },
		})
		if err != nil {
			log.Fatal("Error: ", err)
			log.Fatal("Error: Failed to start the Mongo session")
		}
	}
	return mgoSession.Clone()
}