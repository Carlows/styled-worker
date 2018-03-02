# Styled worker

This Golang program is a worker that will be polling messages from SQS to process images and upload the result to S3!

## Running the program

```
styled-worker -programPath=/home/ubuntu/projects/FastPhotoStyle/demo.py
```

## Compiling for the EC2 ubuntu instance

```
env GOOS=linux GOARCH=arm go build -o build/styled-worker github.com/carlows/styled-worker
```