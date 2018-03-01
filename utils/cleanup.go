package utils

import (
	"log"
	"os"
)

func CleanUpFiles(filePaths []string) {
	for _, filePath := range filePaths {
		err := os.Remove(filePath)
		if err != nil {
			log.Fatal(err)
		}
	}
}
