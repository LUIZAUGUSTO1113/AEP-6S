package main

import "net/http"

const viteOrigin = "http://localhost:5173"

func allowVite(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := w.Header()
		header.Add("Vary", "Origin")
		allowed := r.Header.Get("Origin") == viteOrigin
		if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
			header.Add("Vary", "Access-Control-Request-Method")
			header.Add("Vary", "Access-Control-Request-Headers")
			if allowed {
				header.Set("Access-Control-Allow-Origin", viteOrigin)
				header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
				header.Set("Access-Control-Allow-Headers", "Content-Type")
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if allowed {
			header.Set("Access-Control-Allow-Origin", viteOrigin)
		}
		next.ServeHTTP(w, r)
	})
}
