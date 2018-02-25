package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/stripe/aws-go/aws"
	"github.com/stripe/aws-go/gen/sqs"

	"github.com/carlows/styled-worker/worker"
)

var flagQueue = flag.String("queueName", "example", "specify a queue name")

func Print(msg *sqs.Message) error {
	fmt.Println(*msg.Body)
	time.Sleep(3 * time.Second)
	return nil
}

func main() {
	flag.Parse()

	q, err := worker.NewSQSQueue(
		sqs.New(aws.DetectCreds("", "", ""), "ap-northeast-1", nil),
		*flagQueue,
	)
	if err != nil {
		log.Fatal(err)
	}

	worker.Start(q, worker.HandlerFunc(Print))
}
