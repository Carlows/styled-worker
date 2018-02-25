package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/carlows/styled-worker/worker"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

var flagQueue = flag.String("queueName", "styled-dev-messages", "specify a queue name")

// Print just prints the message from SQS!
func Print(msg *sqs.Message) error {
	fmt.Println(*msg)
	time.Sleep(3 * time.Second)
	return nil
}

func main() {
	flag.Parse()

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

	worker.Start(result.QueueUrl, worker.HandlerFunc(Print), svc)
}
