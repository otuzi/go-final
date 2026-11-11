// Package db отвечает за работу с базой данных SQLite планировщика.
package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite" // регистрация SQL-драйвера "sqlite"
)

// DefaultFile — имя файла базы данных по умолчанию.
const DefaultFile = "scheduler.db"

// schema — SQL-команды для создания таблицы задач и индекса по дате.
// Дата хранится строкой в формате YYYYMMDD (20060102 в Go), поэтому
// сортировка по строке совпадает с сортировкой по дате.
const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    date    CHAR(8)      NOT NULL DEFAULT '',
    title   VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT         NOT NULL DEFAULT '',
    repeat  VARCHAR(128) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`

// db — открытое соединение с базой данных. Заполняется в Init.
var db *sql.DB

// File возвращает путь к файлу базы данных: значение переменной окружения
// TODO_DBFILE, а если она не задана — DefaultFile в текущей директории.
func File() string {
	if f := os.Getenv("TODO_DBFILE"); f != "" {
		return f
	}
	return DefaultFile
}

// Init открывает базу данных по пути dbFile. Если файла ещё не было,
// создаёт таблицу scheduler и индекс по колонке date.
func Init(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := err != nil

	conn, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return fmt.Errorf("открытие базы данных %s: %w", dbFile, err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return fmt.Errorf("подключение к базе данных %s: %w", dbFile, err)
	}

	if install {
		if _, err := conn.Exec(schema); err != nil {
			conn.Close()
			return fmt.Errorf("создание схемы базы данных: %w", err)
		}
	}

	db = conn
	return nil
}

// Close закрывает соединение с базой данных.
func Close() error {
	if db == nil {
		return nil
	}
	return db.Close()
}
