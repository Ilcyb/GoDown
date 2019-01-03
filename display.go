package main

import (
	"fmt"
	"strings"
	"time"
)

type DownloadProcessBar struct {
	Name         string
	Size         DisplaySize
	TotalSize    int64
	CompleteSize int64
	FillChar     string
	Width        int
}

type DisplaySize struct {
	Size float32
	Unit string
}

func displayStatus(bar DownloadProcessBar) {
	fmt.Printf("\n"+
		"%s\t\t\t\t\t\t\t\t\t\t%.2f%s\n", bar.Name, bar.Size.Size, bar.Size.Unit)
}

func displayProcessLine(bar DownloadProcessBar) {
	processWidth := int(float32(float64(bar.CompleteSize)/float64(bar.TotalSize)) * float32(bar.Width))
	process := strings.Repeat(bar.FillChar, processWidth)
	empty := strings.Repeat(" ", bar.Width-processWidth)
	fmt.Printf("[%s>%s]\r", process, empty)
}

func DisplayProcessBar(bar *DownloadProcessBar) {
	displayStatus(*bar)
	for (*bar).CompleteSize < (*bar).TotalSize {
		displayProcessLine(*bar)
		time.Sleep(time.Second)
	}
}

func GetDisplaySizeUnit(bytesSize int64) DisplaySize {
	var displaySize DisplaySize
	if bytesSize < 1024 {
		displaySize.Size = float32(bytesSize)
		displaySize.Unit = "B"
	} else if bytesSize >= 1024 && bytesSize < 1024*1024 {
		displaySize.Size = float32(bytesSize) / 1024
		displaySize.Unit = "Kb"
	} else if bytesSize >= 1024*1024 && bytesSize < 1024*1024*1024 {
		displaySize.Size = float32(bytesSize) / (1024 * 1024)
		displaySize.Unit = "Mb"
	} else if bytesSize >= 1024*1024*1024 && bytesSize < 1024*1024*1024*1024 {
		displaySize.Size = float32(bytesSize) / (1024 * 1024 * 1024)
		displaySize.Unit = "Gb"
	} else {
		panic("resource is too large")
	}

	return displaySize
}
