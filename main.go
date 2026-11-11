package main

import (
	"log"

	"github.com/otuzi/go-final/pkg/db"
	"github.com/otuzi/go-final/pkg/server"
)

func main() {
	if err := db.Init(db.File()); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
