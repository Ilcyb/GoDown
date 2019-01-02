package main

import (
	"flag"
	"fmt"
	"os"
)

type DownloadArguments struct {
	url string
}

func parseArgs() DownloadArguments {
	var url string

	flag.StringVar(&url, "url", "", "the url of the resource you wanted download")

	flag.Parse()

	if url == "" {
		panic("You have to enter a valid url")
	}

	args := DownloadArguments{url}

	return args
}

func main() {
	args := parseArgs()

	//GetFileRange(args.url)
	//os.Exit(0)

	fileName, err := Download(args.url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("%s\n", fileName)
}
