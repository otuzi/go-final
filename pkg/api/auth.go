package api

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// tokenTTL — срок действия JWT-токена. Совпадает со временем жизни куки на фронтенде.
const tokenTTL = 8 * time.Hour

// claims — полезная нагрузка JWT-токена: хэш пароля, по которому был выдан токен.
type claims struct {
	PasswordHash string `json:"hash"`
	jwt.RegisteredClaims
}

// password возвращает пароль из переменной окружения TODO_PASSWORD.
// Пустая строка означает, что аутентификация отключена.
func password() string {
	return os.Getenv("TODO_PASSWORD")
}

// passwordHash возвращает hex-представление SHA-256 от пароля.
// Используется и как полезная нагрузка токена, и как ключ его подписи.
func passwordHash(pass string) string {
	sum := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(sum[:])
}

// makeToken создаёт подписанный JWT-токен для указанного пароля.
func makeToken(pass string) (string, error) {
	hash := passwordHash(pass)
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		PasswordHash: hash,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	})
	return token.SignedString([]byte(hash))
}

// validToken проверяет подпись и срок действия токена, а также что хэш в токене
// совпадает с хэшем текущего пароля. Если пароль изменился, старые токены невалидны.
func validToken(tokenStr, pass string) bool {
	hash := passwordHash(pass)
	var c claims
	_, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неожиданный метод подписи токена")
		}
		return []byte(hash), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(c.PasswordHash), []byte(hash)) == 1
}

// auth — middleware, проверяющий аутентификацию. Если TODO_PASSWORD не задан,
// пропускает запрос без проверки. Иначе требует валидный JWT-токен в куке token
// и при его отсутствии отвечает 401, после чего фронтенд переводит на /login.html.
func auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := password()
		if pass != "" {
			var token string
			if cookie, err := r.Cookie("token"); err == nil {
				token = cookie.Value
			}
			if !validToken(token, pass) {
				writeError(w, http.StatusUnauthorized, "требуется аутентификация")
				return
			}
		}
		next(w, r)
	}
}

// signinHandler обрабатывает POST /api/signin с JSON {"password":"..."}.
// При совпадении пароля с TODO_PASSWORD возвращает {"token":"<JWT>"}, иначе {"error":"..."}.
func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "метод не поддерживается")
		return
	}

	pass := password()
	if pass == "" {
		writeError(w, http.StatusBadRequest, "аутентификация отключена: пароль не задан на сервере")
		return
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "ошибка разбора JSON: "+err.Error())
		return
	}
	if subtle.ConstantTimeCompare([]byte(req.Password), []byte(pass)) != 1 {
		writeError(w, http.StatusUnauthorized, "неверный пароль")
		return
	}

	token, err := makeToken(pass)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "ошибка создания токена: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}
