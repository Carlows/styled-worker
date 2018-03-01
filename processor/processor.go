package processor

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"regexp"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/carlows/styled-worker/utils"
)

type MessageProcessor struct {
	FileUploader *utils.FileUploader
	AWSRegion    string
	BucketName   string
	DB           *utils.DB
	ProgramPath  string
}

func (message *MessageProcessor) Process(msg *sqs.Message) error {
	contentURLAttrib, hasContentURL := msg.MessageAttributes["contentUrl"]
	styleURLAttrib, hasStyleURL := msg.MessageAttributes["styleUrl"]
	recordIDAttrib, hasRecordID := msg.MessageAttributes["recordId"]

	if !hasContentURL || !hasStyleURL || !hasRecordID {
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
		log.Println(err, output)
		return err
	}

	// clean up all files after this function is done
	defer utils.CleanUpFiles([]string{contentPath, stylePath, resultPath})

	cmdOutput := string(output)
	success, err := message.parseCmdOutput(cmdOutput)
	if err != nil {
		return err
	}

	recordID := *recordIDAttrib.StringValue

	// The record failed to update. Let's update the item with an error
	if !success {
		_, err = message.DB.UpdateRecordFailure(recordID)
		if err != nil {
			return err
		}
	}

	key, err := message.FileUploader.UploadFileToS3(resultPath, resultFileName)
	if err != nil {
		return err
	}

	_, err = message.DB.UpdateRecordSuccess(recordID, message.buildS3URL(key))
	if err != nil {
		return err
	}

	fmt.Printf("Sucessfully Updated Item %s", message.buildS3URL(key))

	return nil
}

func (message *MessageProcessor) runCommand(contentPath string, stylePath string, resultPath string) (output []byte, err error) {
	return exec.Command("python", message.ProgramPath, "--content_image_path", contentPath, "--style_image_path", stylePath, "--output_image_path", resultPath).Output()
}

func (message *MessageProcessor) parseCmdOutput(cmdOutput string) (match bool, err error) {
	return regexp.MatchString("(?im)success", cmdOutput)
}

func (message *MessageProcessor) buildS3URL(fileKey string) string {
	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", message.AWSRegion, message.BucketName, fileKey)
}
