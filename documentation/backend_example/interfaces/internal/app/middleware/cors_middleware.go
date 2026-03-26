package middleware

import (
	"github.com/go-chi/cors"
	"net/http"
)

// CORSMiddleware возвращает middleware-функцию, которую можно использовать в chi
func CORSMiddleware() func(next http.Handler) http.Handler {
	// Создаем CORS handler
	corsHandler := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	})

	// Возвращаем middleware-функцию
	return func(next http.Handler) http.Handler {
		return corsHandler(next)
	}
}
