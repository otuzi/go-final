package api

import "testing"

func TestToken(t *testing.T) {
	token, err := makeToken("secret")
	if err != nil {
		t.Fatalf("makeToken: %v", err)
	}
	if !validToken(token, "secret") {
		t.Error("токен для текущего пароля должен быть валидным")
	}
	if validToken(token, "other") {
		t.Error("токен не должен быть валидным после смены пароля")
	}
	if validToken("", "secret") || validToken("garbage", "secret") {
		t.Error("пустой или мусорный токен не должен быть валидным")
	}
}
