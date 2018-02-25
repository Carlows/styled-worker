package worker

import (
	"log"

	"github.com/nabeken/aws-go-sqs/queue"
	"github.com/nabeken/aws-go-sqs/queue/option"
	"github.com/stripe/aws-go/gen/sqs"
)

type HandlerFunc func(msg *sqs.Message) error

func (f HandlerFunc) HandleMessage(msg *sqs.Message) error {
	return f(msg)
}

type Handler interface {
	HandleMessage(msg *sqs.Message) error
}

func Start(q *queue.Queue, h Handler) {
	for {
		log.Println("worker: Start polling")

		messages, err := q.ReceiveMessage(option.MaxNumberOfMessages(10))
		if err != nil {
			log.Println(err)
			continue
		}

		if len(messages) > 0 {
			// Blocks the process until all the messages are processed
			// This is intentional, as we can't run the CLI tool to process images
			// more than once or it will run out of memory :(
			run(q, h, messages)
		}
	}
}

func run(q *queue.Queue, h Handler, messages []sqs.Message) {
	numMessages := len(messages)
	log.Printf("worker: Processing %d messages", numMessages)

	for _, message := range messages {
		log.Println("worker: Processing Message %v", message)

		if err := handleMessage(q, message, h); err != nil {
			log.Println(err)
		}
	}
}

func handleMessage(q *queue.Queue, m *sqs.Message, h Handler) error {
	err := h.HandleMessage(m)
	if err != nil {
		return err
	}

	return q.DeleteMessage(m.ReceiptHandle)
}