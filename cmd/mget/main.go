package main

import (
	"io"
	"log"
	"mget"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		log.Println("mget url")
		return
	}
	url := os.Args[1]
	r, length, err := mget.Get(url)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("length", length)
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Println(err)
		return
	}
}
