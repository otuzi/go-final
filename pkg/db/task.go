package db

// Task — задача планировщика. Используется и при работе с базой данных,
// и при сериализации в JSON для API.
type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// AddTask добавляет задачу в таблицу scheduler и возвращает идентификатор записи.
func AddTask(task *Task) (int64, error) {
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Tasks возвращает до limit ближайших задач, отсортированных по дате.
func Tasks(limit int) ([]*Task, error) {
	return selectTasks(`SELECT id, date, title, comment, repeat FROM scheduler
		ORDER BY date LIMIT ?`, limit)
}

// SearchTasks возвращает до limit задач, в заголовке или комментарии которых
// встречается подстрока search, отсортированных по дате.
func SearchTasks(search string, limit int) ([]*Task, error) {
	pattern := "%" + search + "%"
	return selectTasks(`SELECT id, date, title, comment, repeat FROM scheduler
		WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT ?`, pattern, pattern, limit)
}

// TasksByDate возвращает до limit задач на указанную дату в формате 20060102.
func TasksByDate(date string, limit int) ([]*Task, error) {
	return selectTasks(`SELECT id, date, title, comment, repeat FROM scheduler
		WHERE date = ? ORDER BY id LIMIT ?`, date, limit)
}

// selectTasks выполняет запрос и сканирует строки результата в слайс задач.
// При отсутствии записей возвращает пустой (не nil) слайс.
func selectTasks(query string, args ...any) ([]*Task, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make([]*Task, 0)
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}
