// Package db отвечает за работу с базой данных планировщика.
// Поддерживаются два хранилища: файл SQLite (по умолчанию) и PostgreSQL.
package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрация SQL-драйвера "pgx"
	_ "modernc.org/sqlite"             // регистрация SQL-драйвера "sqlite"
)

// Поддерживаемые драйверы.
const (
	DriverSQLite   = "sqlite"
	DriverPostgres = "pgx"
)

// DefaultFile — имя файла базы данных SQLite по умолчанию.
const DefaultFile = "scheduler.db"

// Config описывает, к какой базе подключаться.
type Config struct {
	// Driver — DriverSQLite или DriverPostgres.
	Driver string
	// DSN — путь к файлу для SQLite или строка подключения для PostgreSQL,
	// например postgres://user:password@host:5432/dbname?sslmode=disable.
	DSN string
}

// ConfigFromEnv собирает конфигурацию из переменных окружения.
// Если задана TODO_POSTGRES_DSN, используется PostgreSQL. Иначе — SQLite
// с файлом из TODO_DBFILE или DefaultFile в текущей директории.
func ConfigFromEnv() Config {
	if dsn := os.Getenv("TODO_POSTGRES_DSN"); dsn != "" {
		return Config{Driver: DriverPostgres, DSN: dsn}
	}
	if f := os.Getenv("TODO_DBFILE"); f != "" {
		return Config{Driver: DriverSQLite, DSN: f}
	}
	return Config{Driver: DriverSQLite, DSN: DefaultFile}
}

// schemaSQLite и schemaPostgres — SQL-команды для создания таблицы задач
// и индекса по дате. Дата хранится строкой в формате YYYYMMDD (20060102 в Go),
// поэтому сортировка по строке совпадает с сортировкой по дате.
const (
	schemaSQLite = `
CREATE TABLE IF NOT EXISTS scheduler (
    id      INTEGER PRIMARY KEY AUTOINCREMENT,
    date    CHAR(8)      NOT NULL DEFAULT '',
    title   VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT         NOT NULL DEFAULT '',
    repeat  VARCHAR(128) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`
	schemaPostgres = `
CREATE TABLE IF NOT EXISTS scheduler (
    id      BIGSERIAL PRIMARY KEY,
    date    CHAR(8)      NOT NULL DEFAULT '',
    title   VARCHAR(255) NOT NULL DEFAULT '',
    comment TEXT         NOT NULL DEFAULT '',
    repeat  VARCHAR(128) NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`
)

// Сколько раз и с каким интервалом повторять подключение к PostgreSQL:
// в docker compose контейнер базы может стартовать чуть позже приложения.
const (
	connectAttempts = 10
	connectDelay    = time.Second
)

// db — открытое соединение с базой данных. Заполняется в Init.
var db *sql.DB

// postgres — true, если открыто соединение с PostgreSQL. Влияет на вид
// плейсхолдеров в запросах, см. rebind.
var postgres bool

// Init открывает базу данных по конфигурации и при необходимости создаёт
// таблицу scheduler с индексом по колонке date.
func Init(cfg Config) error {
	switch cfg.Driver {
	case DriverSQLite:
		return initSQLite(cfg.DSN)
	case DriverPostgres:
		return initPostgres(cfg.DSN)
	default:
		return fmt.Errorf("неизвестный драйвер базы данных %q", cfg.Driver)
	}
}

// initSQLite открывает файл SQLite. Если файла ещё не было, создаёт схему.
func initSQLite(dbFile string) error {
	_, err := os.Stat(dbFile)
	install := err != nil

	conn, err := sql.Open(DriverSQLite, dbFile)
	if err != nil {
		return fmt.Errorf("открытие базы данных %s: %w", dbFile, err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return fmt.Errorf("подключение к базе данных %s: %w", dbFile, err)
	}
	if install {
		if _, err := conn.Exec(schemaSQLite); err != nil {
			conn.Close()
			return fmt.Errorf("создание схемы базы данных: %w", err)
		}
	}

	db, postgres = conn, false
	log.Printf("база данных: SQLite, файл %s", dbFile)
	return nil
}

// initPostgres подключается к PostgreSQL с несколькими попытками
// и создаёт схему, если таблицы ещё нет.
func initPostgres(dsn string) error {
	conn, err := sql.Open(DriverPostgres, dsn)
	if err != nil {
		return fmt.Errorf("открытие подключения к PostgreSQL: %w", err)
	}
	for i := 1; ; i++ {
		if err = conn.Ping(); err == nil {
			break
		}
		if i >= connectAttempts {
			conn.Close()
			return fmt.Errorf("подключение к PostgreSQL: %w", err)
		}
		log.Printf("PostgreSQL недоступен (%v), попытка %d из %d", err, i, connectAttempts)
		time.Sleep(connectDelay)
	}
	if _, err := conn.Exec(schemaPostgres); err != nil {
		conn.Close()
		return fmt.Errorf("создание схемы базы данных: %w", err)
	}

	db, postgres = conn, true
	log.Println("база данных: PostgreSQL")
	return nil
}

// Close закрывает соединение с базой данных.
func Close() error {
	if db == nil {
		return nil
	}
	return db.Close()
}

// rebind приводит плейсхолдеры запроса к виду, который понимает текущий драйвер.
// Запросы в пакете пишутся с "?", PostgreSQL требует "$1", "$2" и так далее.
func rebind(query string) string {
	if !postgres {
		return query
	}
	var b strings.Builder
	b.Grow(len(query) + 8)
	n := 0
	for _, r := range query {
		if r == '?' {
			n++
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(n))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
