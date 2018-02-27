package processor

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/carlows/styled-worker/utils"
)

type MessageProcessor struct {
	FileUploader *utils.FileUploader
	AWSRegion    string
	BucketName   string
}

func (message *MessageProcessor) Process(msg *sqs.Message) error {
	contentURLAttrib, hasContentURL := msg.MessageAttributes["contentUrl"]
	styleURLAttrib, hasStyleURL := msg.MessageAttributes["styleUrl"]

	if !hasContentURL || !hasStyleURL {
		return errors.New("SQS Message has no content or style urls")
	}

	// The naming of these images need to include the extension of the image url
	contentURL := *contentURLAttrib.StringValue
	contentPath, err := utils.DownloadImage("temp", "content.jpg", contentURL)
	if err != nil {
		return err
	}

	styleURL := *styleURLAttrib.StringValue
	stylePath, err := utils.DownloadImage("temp", "style.jpg", styleURL)
	if err != nil {
		return err
	}

	resultFileName := "result.png"
	resultPath := fmt.Sprintf("temp/%s", resultFileName)
	output, err := message.runCommand(contentPath, stylePath, resultPath)
	if err != nil {
		return err
	}

	cmdOutput := string(output)
	success, err := message.parseCmdOutput(cmdOutput)
	if err != nil {
		return err
	}

	// recordId := *msg.MessageAttributes["recordId"].StringValue
	if success {
		key, err := message.FileUploader.UploadFileToS3(resultPath, resultFileName)
		if err != nil {
			return err
		}
		fmt.Printf("File uploaded %s", message.buildS3URL(key))
		// update DynamoDB record with success message
	} else {
		// update DynamoDB record with error message
	}

	// TODO: cleanup of temp folder

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

func (message *MessageProcessor) buildS3URL(fileKey string) string {
	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", message.AWSRegion, message.BucketName, fileKey)
}
