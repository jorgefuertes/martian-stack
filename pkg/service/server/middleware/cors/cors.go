package middleware

import (
	"net/http"
)

func New(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Manejar CORS aquí
		// Por ejemplo, permitir acceso desde un origen específico
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Accept-Language, Accept-Encoding")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// Continuar con el proceso de solicitud
		next.ServeHTTP(w, r)
	})
}
