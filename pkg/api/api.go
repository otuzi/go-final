// Package api содержит обработчики HTTP API планировщика.
package api

import "net/http"

// DateFormat — формат даты, используемый в API и в базе данных (YYYYMMDD).
const DateFormat = "20060102"

// Init регистрирует обработчики API на переданном маршрутизаторе.
// Обработчики задач обёрнуты в middleware auth: при заданной переменной
// окружения TODO_PASSWORD они требуют JWT-токен в куке token.
func Init(mux *http.ServeMux) {
	mux.HandleFunc("/api/signin", signinHandler)
	mux.HandleFunc("/api/nextdate", nextDateHandler)
	mux.HandleFunc("/api/task", auth(taskHandler))
	mux.HandleFunc("/api/task/done", auth(doneTaskHandler))
	mux.HandleFunc("/api/tasks", auth(tasksHandler))
}
