package api

import (
	"net/http"
	"time"

	"github.com/otuzi/go-final/pkg/db"
)

// tasksLimit — максимальное число задач в ответе /api/tasks.
const tasksLimit = 50

// searchDateFormat — формат даты в строке поиска (например, 08.02.2024).
const searchDateFormat = "02.01.2006"

// TasksResp — ответ обработчика /api/tasks.
type TasksResp struct {
	Tasks []*db.Task `json:"tasks"`
}

// tasksHandler обрабатывает GET /api/tasks[?search=<строка или дата 02.01.2006>].
// Без параметра search возвращает ближайшие задачи. Если search — дата,
// возвращает задачи на этот день, иначе ищет подстроку в заголовке и комментарии.
func tasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	var (
		tasks []*db.Task
		err   error
	)
	search := r.FormValue("search")
	switch {
	case search == "":
		tasks, err = db.Tasks(tasksLimit)
	default:
		if d, perr := time.Parse(searchDateFormat, search); perr == nil {
			tasks, err = db.TasksByDate(d.Format(DateFormat), tasksLimit)
		} else {
			tasks, err = db.SearchTasks(search, tasksLimit)
		}
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка получения задач: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, TasksResp{Tasks: tasks})
}
