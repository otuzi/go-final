package api

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/otuzi/go-final/pkg/db"
)

// taskHandler распределяет запросы к /api/task по HTTP-методам.
func taskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTaskHandler(w, r)
	case http.MethodGet:
		getTaskHandler(w, r)
	case http.MethodPut:
		updateTaskHandler(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
	}
}

// getTaskHandler обрабатывает GET /api/task?id=<идентификатор>: возвращает задачу в JSON.
func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.FormValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		writeDBError(w, "ошибка получения задачи", err)
		return
	}
	writeJSON(w, http.StatusOK, task)
}

// updateTaskHandler обрабатывает PUT /api/task: обновляет задачу по id из JSON-тела.
// В ответ возвращает {} либо {"error":"<текст>"}.
func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	task, err := decodeTask(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, err := parseID(task.ID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := db.UpdateTask(task); err != nil {
		writeDBError(w, "ошибка обновления задачи", err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}

// parseID проверяет, что идентификатор задачи задан и является положительным числом.
func parseID(id string) (string, error) {
	if id == "" {
		return "", errors.New("не указан идентификатор задачи")
	}
	if n, err := strconv.ParseInt(id, 10, 64); err != nil || n <= 0 {
		return "", errors.New("некорректный идентификатор задачи: " + id)
	}
	return id, nil
}

// writeDBError отвечает 404, если задача не найдена, иначе 500 с текстом ошибки.
func writeDBError(w http.ResponseWriter, msg string, err error) {
	if errors.Is(err, db.ErrTaskNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, msg+": "+err.Error())
}
