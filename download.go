package main

import (
	"io"
	"net/http"
	"os"
	"strings"
)

func getFileNameFromURL(url string) string {
	hasSlash := strings.Contains(url, "/")
	if !hasSlash {
		return url
	}
	urlSegments := strings.Split(url, "/")
	fileName := urlSegments[len(urlSegments)-1]
	return fileName
}

func Download(url string) (string, error) {
	fileRange, err := GetFileRange(url)
	if err != nil {
		return "", err
	}

	file, err := createFileWithSize(getFileNameFromURL(url), fileRange.FileContentLength)
	if err != nil {
		return "", err
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func createFileWithSize(path string, size uint64) (*os.File, error) {
	fd, err := os.Create(path)
	if err != nil {
		return fd, err
	}

	_, err = fd.Seek(int64(size-1), 0)
	if err != nil {
		return fd, err
	}

	_, err = fd.Write([]byte{0})
	if err != nil {
		return fd, err
	}

	_, err = fd.Seek(0, 0)
	if err != nil {
		return fd, err
	}

	return fd, nil
}
