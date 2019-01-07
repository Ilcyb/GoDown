package HttpDownload

import (
	"testing"
)

func TestGetFileNameFromUrl(t *testing.T) {
	var url, filename string

	url = "https://godown.me/resource1.jpg"
	filename = getFileNameFromURL(url)
	if filename != "resource1.jpg" {
		t.Error()
	}

	url = "https://godown.me/resource1.jpg?q=1"
	filename = getFileNameFromURL(url)
	if filename != "resource1.jpg" {
		t.Error()
	}

	url = "resource1.jpg"
	filename = getFileNameFromURL(url)
	if filename != "resource1.jpg" {
		t.Error()
	}

}
