package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/otuzi/go-final/pkg/db"
)

// addTaskHandler обрабатывает POST /api/task: добавляет задачу в базу данных.
// В ответ возвращает {"id":"<число>"} либо {"error":"<текст>"}.
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		writeError(w, http.StatusBadRequest, "ошибка разбора JSON: "+err.Error())
		return
	}

	if task.Title == "" {
		writeError(w, http.StatusBadRequest, "не указан заголовок задачи")
		return
	}
	if err := checkDate(&task); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	id, err := db.AddTask(&task)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка добавления задачи: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": strconv.FormatInt(id, 10)})
}

// checkDate проверяет и нормализует дату задачи:
//   - пустая дата заменяется на сегодняшнюю;
//   - дата должна быть в формате 20060102;
//   - правило повторения, если указано, должно быть корректным;
//   - если дата раньше сегодняшней, она сдвигается на сегодня (без правила)
//     или на ближайшую дату по правилу повторения.
func checkDate(task *db.Task) error {
	now := time.Now()
	today := now.Format(DateFormat)

	if task.Date == "" {
		task.Date = today
	}
	t, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("дата %q должна быть в формате %s", task.Date, DateFormat)
	}

	var next string
	if task.Repeat != "" {
		next, err = NextDate(now, task.Date, task.Repeat)
		if err != nil {
			return errors.New("некорректное правило повторения: " + err.Error())
		}
	}

	if afterNow(now, t) {
		if task.Repeat == "" {
			task.Date = today
		} else {
			task.Date = next
		}
	}
	return nil
}
