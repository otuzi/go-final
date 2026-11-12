// Package server содержит функцию запуска HTTP-сервера планировщика.
package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/otuzi/go-final/pkg/api"
)

const (
	// defaultPort — порт по умолчанию, если переменная окружения TODO_PORT не задана.
	defaultPort = 7540
	// webDir — директория с файлами фронтенда.
	webDir = "./web"
)

// Run запускает веб-сервер: отдаёт статические файлы фронтенда из директории web
// и регистрирует обработчики API.
// Порт берётся из переменной окружения TODO_PORT, иначе используется 7540.
func Run() error {
	port := getPort()

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(webDir)))
	api.Init(mux)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("сервер запущен на http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}

// getPort возвращает порт из переменной окружения TODO_PORT.
// Если переменная не задана или содержит не число, возвращается порт по умолчанию.
func getPort() int {
	if envPort := os.Getenv("TODO_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p < 65536 {
			return p
		}
		log.Printf("некорректное значение TODO_PORT=%q, используется порт %d", envPort, defaultPort)
	}
	return defaultPort
}
