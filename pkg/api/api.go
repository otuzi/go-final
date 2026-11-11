// Package api содержит обработчики HTTP API планировщика.
package api

import "net/http"

// DateFormat — формат даты, используемый в API и в базе данных (YYYYMMDD).
const DateFormat = "20060102"

// Init регистрирует обработчики API на переданном маршрутизаторе.
func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.HandleFunc("/api/task", taskHandler)
	mux.HandleFunc("/api/tasks", tasksHandler)
}
