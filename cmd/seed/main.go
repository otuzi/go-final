// Команда seed заполняет планировщик демонстрационными задачами через HTTP API.
// Работает с любым хранилищем сервера (SQLite или PostgreSQL), потому что
// не трогает базу напрямую. Если на сервере задан пароль, сначала выполняет вход.
//
// Запуск локально:   go run ./cmd/seed -addr http://localhost:7540
// Запуск в compose:  docker compose --profile seed run --rm seed
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"
)

// task — задача в формате API /api/task.
type task struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

const dateFormat = "20060102"

// sample возвращает набор демонстрационных задач относительно сегодняшней даты.
func sample(now time.Time) []task {
	day := func(n int) string { return now.AddDate(0, 0, n).Format(dateFormat) }
	return []task{
		{day(0), "Проверить почту", "Ответить на важные письма", "d 1"},
		{day(0), "Созвон с командой", "В 11:00, обсудить спринт", "w 1,3,5"},
		{day(1), "Сходить в бассейн", "Взять абонемент", "w 2,4"},
		{day(1), "Купить продукты", "Молоко, хлеб, овощи", ""},
		{day(2), "Оплатить коммуналку", "", "m 25"},
		{day(3), "Прочитать главу книги", "Go: concurrency", "d 3"},
		{day(5), "Встретиться с друзьями", "В 19:00 в кафе", ""},
		{day(7), "Подвести итоги недели", "Записать выводы", "d 7"},
		{day(14), "Сдать отчёт", "До конца дня", ""},
		{day(30), "Проверить страховку", "Продлить, если истекает", "y"},
		{day(0), "Поздравить маму", "Позвонить утром", "m -1"},
		{day(10), "Обновить резюме", "", ""},
	}
}

func main() {
	addr := flag.String("addr", envOr("TODO_ADDR", "http://localhost:7540"), "адрес сервера планировщика")
	password := flag.String("password", os.Getenv("TODO_PASSWORD"), "пароль планировщика, если задан на сервере")
	force := flag.Bool("force", false, "добавить задачи, даже если список уже не пустой")
	wait := flag.Duration("wait", 30*time.Second, "сколько ждать доступности сервера")
	flag.Parse()

	client := newClient()
	if err := waitServer(client, *addr, *wait); err != nil {
		log.Fatal(err)
	}
	if *password != "" {
		if err := signin(client, *addr, *password); err != nil {
			log.Fatal(err)
		}
	}

	count, err := tasksCount(client, *addr)
	if err != nil {
		log.Fatal(err)
	}
	if count > 0 && !*force {
		fmt.Printf("в планировщике уже %d задач, ничего не добавлено (используйте -force)\n", count)
		return
	}

	added := 0
	for _, t := range sample(time.Now()) {
		if err := addTask(client, *addr, t); err != nil {
			log.Printf("не удалось добавить %q: %v", t.Title, err)
			continue
		}
		added++
	}
	fmt.Printf("добавлено задач: %d\n", added)
}

// envOr возвращает значение переменной окружения или def, если она пуста.
func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// newClient создаёт HTTP-клиент с хранилищем кук: в куке token приходит JWT после входа.
func newClient() *http.Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	return &http.Client{Jar: jar, Timeout: 10 * time.Second}
}

// waitServer ждёт, пока сервер начнёт отвечать на запросы главной страницы.
func waitServer(client *http.Client, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		resp, err := client.Get(addr + "/")
		if err == nil {
			resp.Body.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("сервер %s недоступен: %w", addr, err)
		}
		time.Sleep(time.Second)
	}
}

// signin выполняет вход по паролю и сохраняет токен в куку клиента.
func signin(client *http.Client, addr, password string) error {
	var resp struct {
		Token string `json:"token"`
		Error string `json:"error"`
	}
	if err := postJSON(client, addr+"/api/signin", map[string]string{"password": password}, &resp); err != nil {
		return fmt.Errorf("вход: %w", err)
	}
	if resp.Error != "" {
		return fmt.Errorf("вход: %s", resp.Error)
	}
	u, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("адрес сервера: %w", err)
	}
	client.Jar.SetCookies(u, []*http.Cookie{{Name: "token", Value: resp.Token}})
	return nil
}

// tasksCount возвращает число задач в списке /api/tasks.
func tasksCount(client *http.Client, addr string) (int, error) {
	resp, err := client.Get(addr + "/api/tasks")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var body struct {
		Tasks []json.RawMessage `json:"tasks"`
		Error string            `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("список задач: %w", err)
	}
	if body.Error != "" {
		return 0, fmt.Errorf("список задач: %s", body.Error)
	}
	return len(body.Tasks), nil
}

// addTask добавляет одну задачу через POST /api/task.
func addTask(client *http.Client, addr string, t task) error {
	var resp struct {
		ID    string `json:"id"`
		Error string `json:"error"`
	}
	if err := postJSON(client, addr+"/api/task", t, &resp); err != nil {
		return err
	}
	if resp.Error != "" {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

// postJSON отправляет data как JSON и разбирает ответ в out.
func postJSON(client *http.Client, url string, data any, out any) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(out)
}
