package utils

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

type FileUploader struct {
	S3svc      *s3.S3
	BucketName string
}

func (uploader *FileUploader) UploadFileToS3(filePath string, fileName string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size)

	file.Read(buffer)

	filebytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	key := fmt.Sprintf("result-%s.png", uuid.NewV4())

	params := &s3.PutObjectInput{
		Bucket:        aws.String(uploader.BucketName),
		Key:           aws.String(key),
		Body:          filebytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}

	_, err = uploader.S3svc.PutObject(params)
	if err != nil {
		return "", err
	}

	return key, nil
}
