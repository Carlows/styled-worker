package processor

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/carlows/styled-worker/utils"
)

type MessageProcessor struct{}

func (message *MessageProcessor) Process(msg *sqs.Message) error {
	// The naming of these images need to include the extension of the image url
	contentURL := *msg.MessageAttributes["contentUrl"].StringValue
	contentPath, err := utils.DownloadImage("temp", "content.jpg", contentURL)
	if err != nil {
		return err
	}

	styleURL := *msg.MessageAttributes["styleUrl"].StringValue
	stylePath, err := utils.DownloadImage("temp", "style.jpg", styleURL)
	if err != nil {
		return err
	}

	fmt.Println(contentPath, stylePath)

	return nil
}
