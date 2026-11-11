package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// maxDays — максимально допустимый интервал в днях для правила "d".
const maxDays = 400

// NextDate вычисляет следующую дату выполнения задачи.
//
// now — время, от которого ищется ближайшая дата;
// dstart — исходная дата задачи в формате 20060102;
// repeat — правило повторения:
//
//	d <1..400>                      — через указанное число дней;
//	y                               — ежегодно;
//	w <1..7,...>                    — в указанные дни недели (1 — понедельник, 7 — воскресенье);
//	m <дни,...> [месяцы,...]        — в указанные дни месяца (1..31, -1 — последний,
//	                                  -2 — предпоследний), опционально только в указанных месяцах (1..12).
//
// Возвращаемая дата всегда больше и now, и dstart.
func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	if repeat == "" {
		return "", errors.New("правило повторения не указано")
	}

	date, err := time.Parse(DateFormat, dstart)
	if err != nil {
		return "", fmt.Errorf("некорректная дата %q: %w", dstart, err)
	}
	now = truncateDay(now)

	parts := strings.Fields(repeat)
	switch parts[0] {
	case "d":
		return nextByDays(now, date, parts[1:])
	case "y":
		return nextByYear(now, date, parts[1:])
	case "w":
		return nextByWeekdays(now, date, parts[1:])
	case "m":
		return nextByMonthDays(now, date, parts[1:])
	default:
		return "", fmt.Errorf("неподдерживаемый формат правила %q", repeat)
	}
}

// nextByDays обрабатывает правило "d <число>".
func nextByDays(now, date time.Time, args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("правило d: не указан интервал в днях")
	}
	days, err := strconv.Atoi(args[0])
	if err != nil || days < 1 || days > maxDays {
		return "", fmt.Errorf("правило d: интервал должен быть числом от 1 до %d", maxDays)
	}
	for {
		date = date.AddDate(0, 0, days)
		if afterNow(date, now) {
			return date.Format(DateFormat), nil
		}
	}
}

// nextByYear обрабатывает правило "y".
func nextByYear(now, date time.Time, args []string) (string, error) {
	if len(args) != 0 {
		return "", errors.New("правило y: не принимает параметров")
	}
	for {
		date = date.AddDate(1, 0, 0)
		if afterNow(date, now) {
			return date.Format(DateFormat), nil
		}
	}
}

// nextByWeekdays обрабатывает правило "w <дни недели через запятую>".
func nextByWeekdays(now, date time.Time, args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("правило w: не указаны дни недели")
	}
	var weekdays [8]bool
	for _, s := range strings.Split(args[0], ",") {
		wd, err := strconv.Atoi(s)
		if err != nil || wd < 1 || wd > 7 {
			return "", fmt.Errorf("правило w: недопустимый день недели %q", s)
		}
		weekdays[wd] = true
	}

	date = searchStart(now, date)
	for {
		date = date.AddDate(0, 0, 1)
		// time.Weekday: воскресенье = 0, в правиле воскресенье = 7.
		wd := int(date.Weekday())
		if wd == 0 {
			wd = 7
		}
		if weekdays[wd] {
			return date.Format(DateFormat), nil
		}
	}
}

// nextByMonthDays обрабатывает правило "m <дни месяца> [месяцы]".
func nextByMonthDays(now, date time.Time, args []string) (string, error) {
	if len(args) < 1 || len(args) > 2 {
		return "", errors.New("правило m: ожидаются дни месяца и, опционально, месяцы")
	}

	// days[1..31] — обычные дни; lastDays[1] — последний день, lastDays[2] — предпоследний.
	var days [32]bool
	var lastDays [3]bool
	for _, s := range strings.Split(args[0], ",") {
		d, err := strconv.Atoi(s)
		switch {
		case err != nil || d == 0 || d > 31 || d < -2:
			return "", fmt.Errorf("правило m: недопустимый день месяца %q", s)
		case d < 0:
			lastDays[-d] = true
		default:
			days[d] = true
		}
	}

	var months [13]bool
	allMonths := len(args) < 2
	if !allMonths {
		for _, s := range strings.Split(args[1], ",") {
			m, err := strconv.Atoi(s)
			if err != nil || m < 1 || m > 12 {
				return "", fmt.Errorf("правило m: недопустимый месяц %q", s)
			}
			months[m] = true
		}
	}

	date = searchStart(now, date)
	for {
		date = date.AddDate(0, 0, 1)
		if !allMonths && !months[date.Month()] {
			continue
		}
		day := date.Day()
		last := daysInMonth(date)
		if days[day] || (lastDays[1] && day == last) || (lastDays[2] && day == last-1) {
			return date.Format(DateFormat), nil
		}
	}
}

// searchStart возвращает дату, начиная со следующего дня после которой нужно
// искать совпадение: более позднюю из now и date. Результат должен быть
// больше обеих, поэтому даты раньше не имеет смысла проверять.
func searchStart(now, date time.Time) time.Time {
	if afterNow(now, date) {
		return now
	}
	return date
}

// daysInMonth возвращает число дней в месяце указанной даты.
func daysInMonth(t time.Time) int {
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
}

// truncateDay обнуляет время суток, оставляя только дату.
func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// afterNow возвращает true, если дата date строго позже даты now (без учёта времени суток).
func afterNow(date, now time.Time) bool {
	return truncateDay(date).After(truncateDay(now))
}

// nextDateHandler обрабатывает GET /api/nextdate?now=<20060102>&date=<20060102>&repeat=<правило>.
// В ответ возвращает следующую дату в формате 20060102 либо текст ошибки.
func nextDateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	now := time.Now()
	if nowParam := r.FormValue("now"); nowParam != "" {
		parsed, err := time.Parse(DateFormat, nowParam)
		if err != nil {
			http.Error(w, fmt.Sprintf("некорректный параметр now %q", nowParam), http.StatusBadRequest)
			return
		}
		now = parsed
	}

	next, err := NextDate(now, r.FormValue("date"), r.FormValue("repeat"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(next))
}
