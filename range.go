package main

import (
	"net/http"
	"strconv"
	"strings"
)

type FileRange struct {
	SupportRange      bool
	FileContentLength int64
}

/**
检查资源的Range属性支持情况
*/
func GetFileRange(url string) (FileRange, error) {
	fileRange := FileRange{false, 0}

	resp, err := http.Head(url)
	if err != nil {
		return fileRange, err
	}

	header := resp.Header
	// 查看head请求的响应头中是否有accept-ranges字段，如果有该字段则该资源必定支持range属性
	if acceptRange, ok := header["Accept-Ranges"]; ok {
		if strings.Compare(acceptRange[0], "bytes") == 0 {
			fileRange.SupportRange = true
			fileRange.FileContentLength, _ = strconv.ParseInt(header["Content-Length"][0], 10, 64)
		}
	} else {
		// 如果没有accept-ranges也不一定就说明该资源不支持range属性，可以带上range字段访问查看是否支持
		client := &http.Client{}
		req, err := http.NewRequest("HEAD", url, nil)
		if err != nil {
			return fileRange, err
		}

		req.Header.Add("Range", "bytes=0-10")
		anotherResp, err := client.Do(req)
		if err != nil {
			return fileRange, err
		}

		anthorHeader := anotherResp.Header
		littleContentLength, _ := strconv.ParseUint(anthorHeader["Content-Length"][0], 10, 64)
		if littleContentLength == 11 {
			fileRange.SupportRange = true
			fileRange.FileContentLength, _ = strconv.ParseInt(header["Content-Length"][0], 10, 64)
		}
	}
	return fileRange, nil
}
