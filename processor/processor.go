package processor

import (
	"bytes"
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

	fmt.Println("Downloaded both images successfully...")

	resultFileName := "result.png"
	resultPath := fmt.Sprintf("temp/%s", resultFileName)

	// clean up all files after this function is done
	defer utils.CleanUpFiles([]string{contentPath, stylePath, resultPath})

	output, err := message.runCommand(contentPath, stylePath, resultPath)
	if err != nil {
		return err
	}

	fmt.Println("Parsing result output from FastPhotoStyled...")

	cmdOutput := string(output)
	success, err := message.parseCmdOutput(cmdOutput)
	if err != nil {
		return err
	}

	recordID := *recordIDAttrib.StringValue

	// The record failed to update. Let's update the item with an error
	if !success {
		fmt.Println("Failed to style images, updating record with failure")
		_, err = message.DB.UpdateRecordFailure(recordID)
		if err != nil {
			return err
		}
	}

	key, err := message.FileUploader.UploadFileToS3(resultPath, resultFileName)
	if err != nil {
		return err
	}

	fmt.Println("Styling successful! Image was uploaded to S3...")

	_, err = message.DB.UpdateRecordSuccess(recordID, message.buildS3URL(key))
	if err != nil {
		return err
	}

	fmt.Printf("Sucessfully Updated Item %s\n", message.buildS3URL(key))

	return nil
}

func (message *MessageProcessor) runCommand(contentPath string, stylePath string, resultPath string) (string, error) {
	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("python", message.ProgramPath, "--content_image_path", contentPath, "--style_image_path", stylePath, "--output_image_path", resultPath)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if stderr.String() != "" {
		utils.LogError(errors.New(stderr.String()))
	}

	fmt.Println(out.String(), stderr.String())

	return out.String(), err
}

func (message *MessageProcessor) parseCmdOutput(cmdOutput string) (match bool, err error) {
	return regexp.MatchString("(?im)success", cmdOutput)
}

func (message *MessageProcessor) buildS3URL(fileKey string) string {
	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s/%s", message.AWSRegion, message.BucketName, fileKey)
}
