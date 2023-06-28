package main

import (
	"fmt"
	"github.com/dreadl0ck/zeus/zeusutils"
	"log"
)

func main() {
	urlValue := zeusutils.LoadArg("url")
	if urlValue == "" {
		log.Fatal("url value is empty")
	}

	fmt.Println("got url", urlValue)
	fmt.Println("got name", zeusutils.Prompt("name"))
}
