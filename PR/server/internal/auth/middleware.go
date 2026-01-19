// Пакет auth предоставляет функциональность аутентификации для API сервера
package auth

import (
	"context"  // Пакет для работы с контекстом запроса
	"net/http" // Пакет для HTTP-сервера
	"strings"  // Пакет для работы со строками
	"pr-server/internal/config" // Внутренний модуль конфигурации
)

// Тип для ключа контекста
type contextKey string

// Константа для ключа ID агента в контексте
const AgentIDKey contextKey = "agent_id"

// Функция AuthMiddleware создает middleware для проверки токена в заголовке Authorization
// Принимает: конфигурацию сервера
// Возвращает: функцию middleware
// Пропускает аутентификацию для статических файлов и маршрутов веб-интерфейса
func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	// Возвращаем функцию middleware
	return func(next http.Handler) http.Handler {
		// Возвращаем HTTP-обработчик с логикой аутентификации
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Пропускаем аутентификацию для статических файлов и маршрутов веб-интерфейса
			if r.URL.Path == "/" || 
				strings.HasPrefix(r.URL.Path, "/index.html") ||
				strings.HasPrefix(r.URL.Path, "/styles.css") ||
				strings.HasPrefix(r.URL.Path, "/app.js") ||
				strings.HasPrefix(r.URL.Path, "/favicon.ico") ||
				strings.HasPrefix(r.URL.Path, "/assets/") ||
				strings.HasSuffix(r.URL.Path, ".js") ||
				strings.HasSuffix(r.URL.Path, ".css") ||
				strings.HasSuffix(r.URL.Path, ".png") ||
				strings.HasSuffix(r.URL.Path, ".jpg") ||
				strings.HasSuffix(r.URL.Path, ".jpeg") ||
				strings.HasSuffix(r.URL.Path, ".gif") ||
				strings.HasSuffix(r.URL.Path, ".svg") ||
				strings.HasSuffix(r.URL.Path, ".ico") ||
				strings.HasSuffix(r.URL.Path, ".woff") ||
				strings.HasSuffix(r.URL.Path, ".woff2") ||
				strings.HasSuffix(r.URL.Path, ".ttf") ||
				strings.HasSuffix(r.URL.Path, ".eot") {
				// Если путь соответствует статическим файлам, пропускаем аутентификацию
				next.ServeHTTP(w, r)
				return
			}

			// Получаем заголовок Authorization
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Если заголовок отсутствует, возвращаем ошибку 401
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Извлекаем токен из заголовка Authorization
			// Убираем префикс "Bearer "
			token := strings.TrimPrefix(authHeader, "Bearer ")
			// Убираем лишние пробелы
			token = strings.TrimSpace(token)

			// Проверяем, существует ли токен в конфигурации
			if agentID, exists := cfg.Tokens[token]; exists {
				// Если токен действителен, добавляем ID агента в контекст
				ctx := context.WithValue(r.Context(), AgentIDKey, agentID)
				// Продолжаем обработку запроса с обновленным контекстом
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				// Если токен недействителен, возвращаем ошибку 401
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			}
		})
	}
}