package api

import (
	"net/http"
	"time"

	"github.com/otuzi/go-final/pkg/db"
)

// doneTaskHandler обрабатывает POST /api/task/done?id=<идентификатор>.
// Одноразовая задача (без правила повторения) удаляется, периодическая
// переносится на следующую дату по правилу. В ответ возвращает {} либо {"error":"<текст>"}.
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}
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

	if task.Repeat == "" {
		err = db.DeleteTask(id)
	} else {
		var next string
		next, err = NextDate(time.Now(), task.Date, task.Repeat)
		if err != nil {
			writeError(w, http.StatusBadRequest, "некорректное правило повторения: "+err.Error())
			return
		}
		err = db.UpdateDate(next, id)
	}
	if err != nil {
		writeDBError(w, "ошибка выполнения задачи", err)
		return
	}
	writeJSON(w, http.StatusOK, struct{}{})
}
