package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/carlows/styled-worker/processor"
	"github.com/carlows/styled-worker/worker"
)

var flagQueue = flag.String("queueName", "styled-dev-messages", "specify a queue name")

func main() {
	flag.Parse()

	// Create temp folder for images
	_ = os.Mkdir("temp", 0777)

	// Initialize a session in us-west-2 that the SDK will use to load
	// credentials from the shared credentials file ~/.aws/credentials.
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2")},
	)

	// Create a SQS service client.
	svc := sqs.New(sess)

	result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(*flagQueue),
	})

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == sqs.ErrCodeQueueDoesNotExist {
			log.Fatal("Unable to find queue %q.", *flagQueue)
		}
		log.Fatal("Unable to queue %q, %v.", *flagQueue, err)
	}

	fmt.Println("Setting up Worker to listen to:", *result.QueueUrl)

	p := new(processor.MessageProcessor)
	worker.Start(result.QueueUrl, worker.HandlerFunc(p.Process), svc)
}
