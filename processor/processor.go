package processor

import (
	"fmt"
	"os/exec"
	"regexp"

	"github.com/satori/go.uuid"

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

	resultPath := fmt.Sprintf("temp/result-%s.png", uuid.NewV4())
	output, err := message.runCommand(contentPath, stylePath, resultPath)
	if err != nil {
		return err
	}

	cmdOutput := string(output)
	success, err := message.parseCmdOutput(cmdOutput)
	if err != nil {
		return err
	}

	if success {
		// store result in S3
	} else {
		// update DynamoDB record with error message
	}

	return nil
}

func (message *MessageProcessor) runCommand(contentPath string, stylePath string, resultPath string) (output []byte, err error) {
	contentParam := fmt.Sprintf("--content_image_path %s", contentPath)
	styleParam := fmt.Sprintf("--style_image_path %s", stylePath)
	outputParam := fmt.Sprintf("--output_image_path %s", resultPath)
	return exec.Command("python", "demo.py", contentParam, styleParam, outputParam).Output()
}

func (message *MessageProcessor) parseCmdOutput(cmdOutput string) (match bool, err error) {
	return regexp.MatchString("(?im)success", cmdOutput)
}
