package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/carlows/styled-worker/utils"
	raven "github.com/getsentry/raven-go"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/carlows/styled-worker/processor"
	"github.com/carlows/styled-worker/worker"
)

var flagQueue = flag.String("queueName", "styled-dev-messages", "specify a queue name")
var flagBucket = flag.String("bucketName", "styled-dev-test", "specify a bucket name to upload files to")
var flagRegion = flag.String("region", "us-west-2", "specify s3 region")
var flagRecordsTable = flag.String("tableName", "styled-dev", "specify dynamodb table")
var flagProgramPath = flag.String("programPath", "demo.py", "specify path for python demo")

func main() {
	// Configure Raven
	raven.SetDSN("https://e724268551344494aeec3cbf310a23a2:cde51d041ae54fc4941f97c2cb08fd62@sentry.io/297039")

	flag.Parse()

	// Create temp folder for images
	_ = os.Mkdir("temp", 0777)

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(*flagRegion)},
	)

	// Create a SQS service client.
	sqssvc := sqs.New(sess)
	// Create S3 service client.
	s3svc := s3.New(sess)
	// Create DynamoDB service client.
	dynamodbsvc := dynamodb.New(sess)

	result, err := sqssvc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(*flagQueue),
	})

	if err != nil {
		raven.CaptureErrorAndWait(err, nil)
		log.Fatal("Unable to queue %q, %v.", *flagQueue, err)
	}

	fmt.Println("Setting up Worker to listen to:", *result.QueueUrl)

	uploader := &utils.FileUploader{
		S3svc:      s3svc,
		BucketName: *flagBucket,
	}
	db := &utils.DB{
		DynamoDB:  dynamodbsvc,
		TableName: *flagRecordsTable,
	}
	processor := &processor.MessageProcessor{
		FileUploader: uploader,
		AWSRegion:    *flagRegion,
		BucketName:   *flagBucket,
		DB:           db,
		ProgramPath:  *flagProgramPath,
	}

	worker.Start(result.QueueUrl, worker.HandlerFunc(processor.Process), sqssvc)
}
