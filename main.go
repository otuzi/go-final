package main

import (
	"log"

	"github.com/otuzi/go-final/pkg/server"
)

func main() {
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
