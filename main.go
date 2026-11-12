package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/otuzi/go-final/pkg/db"
	"github.com/otuzi/go-final/pkg/server"
)

func main() {
	// Переменные из .env подгружаются, если файл есть. Уже заданные в окружении
	// переменные имеют приоритет и файлом не перезаписываются.
	if err := godotenv.Load(); err == nil {
		log.Println("загружены переменные окружения из .env")
	}

	if err := db.Init(db.ConfigFromEnv()); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
