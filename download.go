package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

const (
	//MaxMemoryUsage 最大内存占用
	MaxMemoryUsage = 1024 * 1024 * 100

	//并发下载数量
	DownloadThreadNum = 8
)

func getFileNameFromURL(resUrl string) string {
	resUrl, _ = url.QueryUnescape(resUrl)
	hasSlash := strings.Contains(resUrl, "/")
	if !hasSlash {
		return resUrl
	}
	urlSegments := strings.Split(resUrl, "/")
	fileName := urlSegments[len(urlSegments)-1]

	return fileName
}

func Download(resUrl string) (string, error) {
	fileRange, err := GetFileRange(resUrl)
	if err != nil {
		return "", err
	}

	file, err := createFileWithSize(getFileNameFromURL(resUrl), fileRange.FileContentLength)
	if err != nil {
		return "", err
	}
	defer file.Close()

	parallelDownload(resUrl, file, fileRange)

	//resp, err := http.Get(resUrl)
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

func parallelDownload(resUrl string, fd *os.File, fileRange FileRange) {

	needDownloadDatas := getGoroutineDatas(fileRange)
	downloadChain := make(chan downloadGoroutineData, DownloadThreadNum)
	downloadedChain := make(chan downloadGoroutineData, DownloadThreadNum)

	downloadCompleteWait := sync.WaitGroup{}
	downloadCompleteWait.Add(len(needDownloadDatas))

	fdmu := sync.Mutex{}

	// 利用通道将同时发起的下载goroutine最多只有 DownloadThreadNum 个
	go func() {
		for i := 0; i < len(needDownloadDatas); i++ {
			downloadChain <- needDownloadDatas[i]
		}
	}()

	// 持续的将处于通道中的下载任务取出进行处理
	go func() {
		for {
			downloadTask := <-downloadChain
			go func() {
				downloadTask, err := downloadGoroutine(resUrl, downloadTask)
				if err != nil {
					panic(err.Error())
				}
				downloadedChain <- downloadTask
			}()
		}
	}()

	go func() {
		for {
			downloadedTask := <-downloadedChain
			go func() {
				fdmu.Lock()
				offset, err := fd.Seek(downloadedTask.begin, 0)
				if err != nil {
					panic(err.Error())
				}
				n, err := fd.Write(downloadedTask.content)
				if err != nil {
					panic(err.Error())
				}
				fmt.Printf("begin:%d end:%d content-length:%d offset:%d success write:%d bytes\n", downloadedTask.begin, downloadedTask.end, len(downloadedTask.content), offset, n)
				downloadCompleteWait.Done()
				fdmu.Unlock()
			}()
		}
	}()

	downloadCompleteWait.Wait()
}

func downloadGoroutine(resUrl string, data downloadGoroutineData) (downloadGoroutineData, error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", resUrl, nil)
	if err != nil {
		return data, errors.New(fmt.Sprintf("occur error while create http client:%s", err.Error()))
	}

	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", data.begin, data.end))

	resp, err := client.Do(req)
	if err != nil {
		return data, errors.New(fmt.Sprintf("occur error while download data:%s", err.Error()))
	}

	data.content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, errors.New(fmt.Sprintf("occur error while read data from response:%s", err.Error()))
	}

	if len(data.content) != int(data.end-data.begin+1) {
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

	for i = 0; i <= DownloadThreadNum; i++ {
		if i == DownloadThreadNum {
			begin = i * pSize
			end = fileRange.FileContentLength - 1
		} else {
			begin = i * pSize
			end = begin + pSize - 1
		}
		data := downloadGoroutineData{begin, end, false, []byte{}}
		downloadGoroutineDatas = append(downloadGoroutineDatas, data)
	}

	return downloadGoroutineDatas
}
