package worker

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/carlows/styled-worker/utils"
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
			MaxNumberOfMessages:   aws.Int64(10),
			MessageAttributeNames: aws.StringSlice([]string{"contentUrl", "styleUrl", "recordId"}),
			WaitTimeSeconds:       aws.Int64(20),
		})
		if err != nil {
			utils.LogError(err)
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
	log.Println("worker: Processing %d messages", numMessages)

	for _, message := range messages {
		if err := handleMessage(queueUrl, message, h, svc); err != nil {
			utils.LogError(err)
		}
	}
}

func handleMessage(queueUrl *string, m *sqs.Message, h Handler, svc *sqs.SQS) error {
	if err := h.HandleMessage(m); err != nil {
		utils.LogError(err)
	}

	_, err := svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      queueUrl,
		ReceiptHandle: m.ReceiptHandle,
	})

	if err != nil {
		// Stop everything if we cannot keep deleting messages
		// Otherwise we will keep processing the same messages over and over
		utils.LogErrorAndDie(err)
	}

	return nil
}
