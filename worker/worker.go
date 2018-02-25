package worker

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type HandlerFunc func(msg *sqs.Message) error

func (f HandlerFunc) HandleMessage(msg *sqs.Message) error {
	return f(msg)
}

type Handler interface {
	HandleMessage(msg *sqs.Message) error
}

func Start(queueUrl *string, h Handler, svc *sqs.SQS) {
	log.Println("worker: Start polling")
	for {
		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl: queueUrl,
			AttributeNames: aws.StringSlice([]string{
				"SentTimestamp",
			}),
			MaxNumberOfMessages: aws.Int64(10),
			MessageAttributeNames: aws.StringSlice([]string{
				"All",
			}),
			WaitTimeSeconds: aws.Int64(20),
		})
		if err != nil {
			log.Println(err)
			continue
		}

		if len(result.Messages) > 0 {
			// Blocks the process until all the messages are processed
			// This is intentional, as we can't run the CLI tool to process images
			// more than once or it will run out of memory :(
			run(queueUrl, h, result.Messages, svc)
		}
	}
}

func run(queueUrl *string, h Handler, messages []*sqs.Message, svc *sqs.SQS) {
	numMessages := len(messages)
	log.Printf("worker: Processing %d messages", numMessages)

	for _, message := range messages {
		if err := handleMessage(queueUrl, message, h, svc); err != nil {
			log.Println(err)
		}
	}
}

func handleMessage(queueUrl *string, m *sqs.Message, h Handler, svc *sqs.SQS) error {
	err := h.HandleMessage(m)
	if err != nil {
		return err
	}

	_, delErr := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      queueUrl,
		ReceiptHandle: m.ReceiptHandle,
	})

	if delErr != nil {
		// Stop everything if we cannot keep deleting messages
		// Otherwise we will keep processing the same messages over and over
		log.Fatal(err)
	}

	return nil
}
