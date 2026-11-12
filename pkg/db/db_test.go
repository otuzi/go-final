package db

import "testing"

func TestRebind(t *testing.T) {
	const q = "SELECT id FROM scheduler WHERE title LIKE ? OR comment LIKE ? LIMIT ?"

	postgres = false
	if got := rebind(q); got != q {
		t.Errorf("для SQLite запрос должен остаться без изменений, получено %q", got)
	}

	postgres = true
	defer func() { postgres = false }()
	want := "SELECT id FROM scheduler WHERE title LIKE $1 OR comment LIKE $2 LIMIT $3"
	if got := rebind(q); got != want {
		t.Errorf("rebind = %q, ожидалось %q", got, want)
	}
}
