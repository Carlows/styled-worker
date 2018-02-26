package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/satori/go.uuid"
)

// DownloadImage - Downloads an image to the filesystem and returns its path
func DownloadImage(folder string, name string, url string) (string, error) {
	path := fmt.Sprintf("%s/%s-%s", folder, uuid.NewV4(), name)
	img, _ := os.Create(path)
	defer img.Close()

	resp, err := http.Get(url)
	if err != nil {
		return path, err
	}
	defer resp.Body.Close()

	_, err = io.Copy(img, resp.Body)
	if err != nil {
		return path, err
	}

	return path, nil
}

func DeleteImage(path string) error {
	return os.Remove(path)
}
