package HttpDownload

import (
	"fmt"
	"os"
	"testing"
)

func TestGetFileNameFromUrl(t *testing.T) {
	var url, filename string

	url = "https://godown.me/resource1.jpg"
	filename = getFileNameFromURL(url)
	if filename != "resource1.jpg" {
		t.Error(filename)
	}

	url = "https://godown.me/resource1.jpg?q=1"
	filename = getFileNameFromURL(url)
	if filename != "resource1.jpg" {
		t.Error(filename)
	}

	url = "resource1.jpg"
	filename = getFileNameFromURL(url)
	if filename != "resource1.jpg" {
		t.Error(filename)
	}

	url = "https://godown.me/this%20is%20a%20resource.jpg"
	filename = getFileNameFromURL(url)
	if filename != "this is a resource.jpg" {
		t.Error(filename)
	}

	url = "https://godown.me/this%20is%20a%20resource.jpg?q=1"
	filename = getFileNameFromURL(url)
	if filename != "this is a resource.jpg" {
		t.Error(filename)
	}
}

func TestCreateFileWithSize(t *testing.T) {
	var size int64
	var fileName string
	var file *os.File

	size = 1024
	fileName = "test.tmp"
	file, err := createFileWithSize(fileName, size)
	defer func() {
		file.Close()
		os.Remove(fileName)
	}()
	if err != nil {
		t.Error(err.Error())
	}

	stat, err := os.Stat(fileName)
	if err != nil {
		t.Error(err)
	}

	if stat.Name() != fileName {
		t.Error(stat.Name())
	}

	if stat.Size() != size {
		t.Error(stat.Size())
	}

}

func TestDownloadGoroutine(t *testing.T) {
	var resUrl string
	var goroutineData downloadGoroutineData

	resUrl = "https://images.pexels.com/photos/1553961/pexels-photo-1553961.jpeg?auto=compress&cs=tinysrgb&dpr=2&w=500"
	goroutineData = downloadGoroutineData{100, 110, false, []byte{}, 0}
	goroutineData, err := downloadGoroutine(resUrl, goroutineData)
	if err != nil {
		t.Error(err.Error())
	}

	if goroutineData.complete != true || len(goroutineData.content) != (110-100+1) {
		t.Error(goroutineData)
	}
}

func TestGetGoroutineDatas(t *testing.T) {
	var fileRange FileRange
	var httpDownloadStrategy int
	var goroutineDatas []downloadGoroutineData
	var fileSize int64

	fileSize = 102488

	httpDownloadStrategy = HttpLargeFileParallelDownload
	fileRange = FileRange{false, fileSize}
	goroutineDatas = getGoroutineDatas(fileRange, httpDownloadStrategy)
	if len(goroutineDatas) != 1 || goroutineDatas[0].begin != 0 || goroutineDatas[0].end != fileSize-1 {
		t.Error(goroutineDatas)
	}

	httpDownloadStrategy = HttpDownload
	fileRange = FileRange{true, fileSize}
	goroutineDatas = getGoroutineDatas(fileRange, httpDownloadStrategy)
	if len(goroutineDatas) != 1 || goroutineDatas[0].begin != 0 || goroutineDatas[0].end != fileSize-1 {
		t.Error(goroutineDatas)
	}

	httpDownloadStrategy = HttpSmallFileParallelDownload
	fileRange = FileRange{true, fileSize}
	goroutineDatas = getGoroutineDatas(fileRange, httpDownloadStrategy)
	lastData := goroutineDatas[len(goroutineDatas)-1]
	if len(goroutineDatas) != DownloadThreadNum+1 || goroutineDatas[0].begin != 0 || goroutineDatas[len(goroutineDatas)-1].end != fileSize-1 {
		t.Error(fmt.Sprintf("len(goroutineDatas):%d, goroutineDatas[0].begin:%d goroutineDatas[len(goroutineDatas)-1].end:%d fileSize:%d",
			len(goroutineDatas), goroutineDatas[0].begin, goroutineDatas[len(goroutineDatas)-1].end, fileSize))
	}
	if fileSize%DownloadThreadNum == 0 {
		if (lastData.end - lastData.begin + 1) != fileSize/DownloadThreadNum {
			t.Error(fmt.Sprintf("last data length:%d, except length:%d", lastData.end-lastData.begin+1, fileSize/DownloadThreadNum))
		}
	} else {
		if (lastData.end - lastData.begin + 1) != fileSize%DownloadThreadNum {
			t.Error(fmt.Sprintf("last data length:%d, except length:%d", lastData.end-lastData.begin+1, fileSize%DownloadThreadNum))
		}
	}

	httpDownloadStrategy = HttpLargeFileParallelDownload
	fileRange = FileRange{true, fileSize}
	goroutineDatas = getGoroutineDatas(fileRange, httpDownloadStrategy)
	if fileRange.FileContentLength%DownloadSizePerThread == 0 {
		if len(goroutineDatas) != int(fileRange.FileContentLength/DownloadSizePerThread) {
			t.Error(fmt.Sprintf("data length:%d, except length:%d", len(goroutineDatas), fileRange.FileContentLength/DownloadSizePerThread))
		}
	} else {
		lastData = goroutineDatas[len(goroutineDatas)-1]
		if len(goroutineDatas) != int(fileRange.FileContentLength/DownloadSizePerThread)+1 ||
			(lastData.end-lastData.begin+1) != fileRange.FileContentLength%DownloadSizePerThread {
			t.Error(fmt.Sprintf("last data length:%d, except length:%d", lastData.end-lastData.begin+1, fileSize%DownloadThreadNum))
		}
	}
}
