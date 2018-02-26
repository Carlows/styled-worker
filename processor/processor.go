package processor

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/carlows/styled-worker/utils"
)

type MessageProcessor struct{}

func (message *MessageProcessor) Process(msg *sqs.Message) error {
	contentURL := *msg.MessageAttributes["contentUrl"].StringValue
	styleURL := *msg.MessageAttributes["styleUrl"].StringValue

	contentPath, err := utils.DownloadImage("temp", "content.jpg", contentURL)
	if err != nil {
		return err
	}

	// The naming of these images need to include the extension of the image url
	stylePath, err := utils.DownloadImage("temp", "style.jpg", styleURL)
	if err != nil {
		return err
	}

	fmt.Println(contentPath, stylePath)

	return nil
}
