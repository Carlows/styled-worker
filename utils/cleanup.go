package utils

import (
	"os"
)

func CleanUpFiles(filePaths []string) {
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); err == nil {
			err := os.Remove(filePath)
			if err != nil {
				LogErrorAndDie(err)
			}
		}
	}
}
