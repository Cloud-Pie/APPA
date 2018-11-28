package run

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"log"
	"io"
)

func createBucketMinio(s3BucketName string) bool{
	bucket := aws.String(s3BucketName)

	sessionAWS := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AWSConfig.AwsAccessKeyId, AWSConfig.AwsSecretAccessKey, ""),
		Region:      aws.String(AWSConfig.Region),
		Endpoint:    aws.String("http://minio:9000"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}))
	// Create S3 service client
	s3Client := s3.New(sessionAWS)

	cparams := &s3.CreateBucketInput{
		Bucket: bucket, // Required
	}

	// Create a new bucket using the CreateBucket call.
	_, err := s3Client.CreateBucket(cparams)
	if err != nil {
		// Message from an error.
		log.Println(err.Error())
		return false
	}
	return true
}
func uploadObjectBucket(s3BucketName string, uploadObject io.ReadSeeker, keyName string){
	bucket := aws.String(s3BucketName)
	key := aws.String(keyName)
	sessionAWS := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(AWSConfig.AwsAccessKeyId, AWSConfig.AwsSecretAccessKey, ""),
		Region:      aws.String(AWSConfig.Region),
		Endpoint:    aws.String("http://minio:9000"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}))
	// Create S3 service client
	s3Client := s3.New(sessionAWS)

	// Upload a new object "testobject" with the string "Hello World!" to our "newbucket".
	_, err := s3Client.PutObject(&s3.PutObjectInput{
		Body:   uploadObject,
		Bucket: bucket,
		Key:    key,
	})

	if err != nil {
		fmt.Printf("Failed to upload data to %s/%s, %s\n", *bucket, *key, err.Error())
		return
	}
	fmt.Printf("Successfully created bucket %s and uploaded data with key %s\n", *bucket, *key)
}


/*
for docker compose file
  minio:
    image: minio/minio
    container_name: 'minio_server'
    volumes:
      - ./miniodata:/data
    ports:
      - "9000:9000"
 */