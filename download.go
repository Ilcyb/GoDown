package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	//MaxMemoryUsage 最大内存占用
	MaxMemoryUsage = 1024 * 1024 * 100

	//并发下载数量
	DownloadThreadNum = 8
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

	parallelDownload(url, file, fileRange)

	//resp, err := http.Get(url)
	//if err != nil {
	//	return "", err
	//}
	//defer resp.Body.Close()
	//
	//_, err = io.Copy(file, resp.Body)
	//if err != nil {
	//	return "", err
	//}

	return file.Name(), nil
}

func createFileWithSize(path string, size int64) (*os.File, error) {
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

type downloadGoroutineData struct {
	begin    int64
	end      int64
	complete bool
	content  []byte
}

func parallelDownload(url string, fd *os.File, fileRange FileRange) {
	needDownloadDatas := getGoroutineDatas(fileRange)
	downloadChain := make(chan downloadGoroutineData, len(needDownloadDatas))
	downloadedChain := make(chan downloadGoroutineData, DownloadThreadNum)

	for i := 0; i < len(needDownloadDatas); i++ {
		downloadChain <- needDownloadDatas[i]
	}

	for i := 0; i < DownloadThreadNum; i++ {
		go func() {
			data := <-downloadChain
			data, _ = downloadGoroutine(url, data)
			downloadedChain <- data
		}()
	}

	for i := 0; i < DownloadThreadNum; i++ {
		go func() {
			data := <-downloadedChain
			fd.Seek(data.begin, 0)
			fd.Write(data.content)
		}()
	}
}

func downloadGoroutine(url string, data downloadGoroutineData) (downloadGoroutineData, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return data, errors.New(fmt.Sprintf("occur error while create http client:%s", err.Error()))
	}

	req.Header.Add("Range", fmt.Sprintf("%d-%d", data.begin, data.end))

	resp, err := client.Do(req)
	if err != nil {
		return data, errors.New(fmt.Sprintf("occur error while download data:%s", err.Error()))
	}

	n, err := resp.Body.Read(data.content)
	if err != nil {
		return data, errors.New(fmt.Sprintf("occur error while read data from response:%s", err.Error()))
	}

	if n != int(data.end-data.begin+1) {
		return data, errors.New("the length of download data is not same with except")
	}

	data.complete = true

	return data, nil
}

func getGoroutineDatas(fileRange FileRange) []downloadGoroutineData {
	var downloadGoroutineDatas []downloadGoroutineData
	var begin int64
	var end int64
	var i int64

	pSize := fileRange.FileContentLength / DownloadThreadNum

	for i = 0; i <= DownloadThreadNum; i += pSize {
		if i == DownloadThreadNum {
			begin = i
			end = fileRange.FileContentLength
		} else {
			begin = i
			end = begin + pSize - 1
		}
		data := downloadGoroutineData{begin, end, false, make([]byte, end-begin+1)}
		downloadGoroutineDatas = append(downloadGoroutineDatas, data)
	}

	return downloadGoroutineDatas
}
