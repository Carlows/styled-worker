package utils

import (
	"log"

	"github.com/getsentry/raven-go"
)

func LogError(err error) {
	raven.CaptureError(err, nil)
	log.Println(err)
}

func LogErrorAndDie(err error) {
	raven.CaptureErrorAndWait(err, nil)
	log.Fatalln(err)
}
